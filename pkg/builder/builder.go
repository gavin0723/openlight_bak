// Author: lipixun
// File Name: builder.go
// Description:

package builder

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/builder/buildcontext"
	"github.com/ops-openlight/openlight/pkg/builder/targetbuilder"
	"github.com/ops-openlight/openlight/pkg/repository"
	"github.com/ops-openlight/openlight/pkg/utils"
)

// Builder implements the build functions
//	A builder contains full context data of a build.
//	Call build methods (e.g. BuildTarget, BuildTargetDependencies) will share the same context data
//	(such as cache data, build history, build tag, etc...)
type Builder struct {
	BuildContext
}

// NewBuilder creates a new Builder
func NewBuilder(root *repository.LocalRepository, verbose bool) (*Builder, error) {
	if root == nil {
		return nil, errors.New("Require root")
	}
	// Done
	tag, err := utils.NewTag()
	if err != nil {
		return nil, fmt.Errorf("Failed to create tag: %v", err)
	}
	return &Builder{BuildContext{verbose: verbose, rootRepo: root, tag: tag}}, nil
}

// NewBuilderFromPath creates a new Builder from path
func NewBuilderFromPath(path string, verbose bool) (*Builder, error) {
	repo, err := repository.NewLocalRepository(path)
	if err != nil {
		return nil, err
	}
	return NewBuilder(repo, verbose)
}

// BuildTargetOption defines the options of building target
type BuildTargetOption interface {
	BuildTargetDependencyOption

	set(options *_BuildTargetOption)
}

type _BuildTargetOption struct {
	noBuild bool
	// Target options
	goBinaryTarget targetbuilder.GoBinaryTargetBuildOptions
}

func (o *_BuildTargetOption) GetNoBuild() bool {
	if o == nil {
		return false
	}
	return o.noBuild
}

func (o *_BuildTargetOption) GetGoBinaryTarget() targetbuilder.GoBinaryTargetBuildOptions {
	if o == nil {
		return targetbuilder.GoBinaryTargetBuildOptions{}
	}
	return o.goBinaryTarget
}

func (o *_BuildTargetOption) String() string {
	return fmt.Sprintf("NoBuild [%v]", o.GetNoBuild())
}

// BuildTarget build a target
func (builder *Builder) BuildTarget(name string, options ...BuildTargetOption) error {
	var _targetOptions _BuildTargetOption
	var _depOptions _BuildTargetDependencyOptions
	for _, option := range options {
		option.set(&_targetOptions)
		option.setdep(&_depOptions)
	}
	_depOptions.BuildTarget = true
	// Get Target
	pkg, err := builder.RootRepository().RootPackage()
	if err != nil {
		return err
	}
	target := pkg.GetTarget(name)
	if target == nil {
		return fmt.Errorf("Unknown target [%v]", name)
	}
	// Build dep
	return builder.buildTarget(target, &_targetOptions, &_depOptions)
}

func (builder *Builder) buildTarget(target *repository.Target, targetOptions *_BuildTargetOption, depOptions *_BuildTargetDependencyOptions) error {
	if target == nil {
		return errors.New("Require target")
	}

	log.Infof("Build target [%v]", target.Name())
	log.Debugf("Builder.buildTarget: Build target [%v]. Options: %v DepOptions: %v", target.Name(), targetOptions.String(), depOptions.String())
	// Check building chain
	if builder.IsTargetBuilding(target) {
		log.Errorln("Circle building dependency found")
		return errors.New("Circle building dependency found")
	}

	// Set target building state
	builder.setTargetBuilding(target)
	defer func() { builder.unsetTargetBuilding(target) }()

	// Build the dependencies of this target
	if err := builder.buildTargetDependencies(target, targetOptions, depOptions); err != nil {
		return err
	}

	// Build target
	if targetOptions.GetNoBuild() {
		return nil
	}
	var err error
	var result *buildcontext.TargetBuildResult
	if spec := target.Spec().GetCommand(); spec != nil {
		result, err = builder.buildCommandTarget(target, spec, targetOptions)
	} else if spec := target.Spec().GetGoBinary(); spec != nil {
		result, err = builder.buildGoBinaryTarget(target, spec, targetOptions)
	} else if spec := target.Spec().GetPythonLib(); spec != nil {
		result, err = builder.buildPythonLibTarget(target, spec, targetOptions)
	} else if spec := target.Spec().GetDockerImage(); spec != nil {
		result, err = builder.buildDockerImageTarget(target, spec, targetOptions)
	} else {
		// A target without specific target. This kind of target is used as an "aggregation" of other targets
		result = new(buildcontext.TargetBuildResult)
	}
	if err != nil {
		return err
	}
	if result == nil {
		log.Errorln("No build result returned")
		return errors.New("No build result returned")
	}
	if result.Artifact() != nil {
		log.Infoln("Generated artifact:", result.Artifact().String())
	}
	builder.addBuildResult(target, result)

	// Done
	return nil
}

func (builder *Builder) buildCommandTarget(target *repository.Target, spec *pbSpec.CommandTarget, options *_BuildTargetOption) (*buildcontext.TargetBuildResult, error) {
	b, err := targetbuilder.NewCommandTargetBuilder(target, spec)
	if err != nil {
		return nil, err
	}
	art, err := b.Build(builder)
	if err != nil {
		return nil, err
	}
	return buildcontext.NewTargetBuildResult(art), nil
}

func (builder *Builder) buildGoBinaryTarget(target *repository.Target, spec *pbSpec.GoBinaryTarget, options *_BuildTargetOption) (*buildcontext.TargetBuildResult, error) {
	b, err := targetbuilder.NewGoBinaryTargetBuilder(target, spec, options.GetGoBinaryTarget())
	if err != nil {
		return nil, err
	}
	art, err := b.Build(builder)
	if err != nil {
		return nil, err
	}
	return buildcontext.NewTargetBuildResult(art), nil
}

func (builder *Builder) buildPythonLibTarget(target *repository.Target, spec *pbSpec.PythonLibTarget, options *_BuildTargetOption) (*buildcontext.TargetBuildResult, error) {
	b, err := targetbuilder.NewPythonLibTargetBuilder(target, spec)
	if err != nil {
		return nil, err
	}
	art, err := b.Build(builder)
	if err != nil {
		return nil, err
	}
	return buildcontext.NewTargetBuildResult(art), nil
}

func (builder *Builder) buildDockerImageTarget(target *repository.Target, spec *pbSpec.DockerImageTarget, options *_BuildTargetOption) (*buildcontext.TargetBuildResult, error) {
	b, err := targetbuilder.NewDockerImageTargetBuilder(target, spec)
	if err != nil {
		return nil, err
	}
	art, err := b.Build(builder)
	if err != nil {
		return nil, err
	}
	return buildcontext.NewTargetBuildResult(art), nil
}

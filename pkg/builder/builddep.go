// Author: lipixun
// File Name: builddep.go
// Description:
//
//	Build the dependencies of a target
//

package builder

import (
	"errors"
	"fmt"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/repository"
)

// BuildTargetDependencyOption defines the options of building dependencies of target
type BuildTargetDependencyOption interface {
	setdep(options *_BuildTargetDependencyOptions)
}

type _BuildTargetDependencyOptions struct {
	NoPip       bool // Do not build pip dependencies
	NoGo        bool // Do not build go dependencies
	NoTarget    bool // Do not build target dependencies
	Update      bool // Update the dependencies or not
	BuildTarget bool // Build target itself
}

func (o *_BuildTargetDependencyOptions) GetNoPip() bool {
	if o == nil {
		return false
	}
	return o.NoPip
}

func (o *_BuildTargetDependencyOptions) GetNoGo() bool {
	if o == nil {
		return false
	}
	return o.NoGo
}

func (o *_BuildTargetDependencyOptions) GetNoTarget() bool {
	if o == nil {
		return false
	}
	return o.NoTarget
}

func (o *_BuildTargetDependencyOptions) GetUpdate() bool {
	if o == nil {
		return false
	}
	return o.Update
}

func (o *_BuildTargetDependencyOptions) String() string {
	return fmt.Sprintf("NoPip [%v] NoGo [%v] NoTarget [%v] Update [%v]",
		o.GetNoPip(),
		o.GetNoGo(),
		o.GetNoTarget(),
		o.GetUpdate(),
	)
}

// BuildTargetDependencies build a target's dependencies
func (builder *Builder) BuildTargetDependencies(name string, options ...BuildTargetOption) error {
	var _targetOptions _BuildTargetOption
	var _options _BuildTargetDependencyOptions
	for _, option := range options {
		option.set(&_targetOptions)
		option.setdep(&_options)
	}
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
	return builder.buildTargetDependencies(target, &_targetOptions, &_options)
}

// DependencyCache stores the installed dependencies (except for the target dependency)
type _DependencyCache struct {
	goPackages map[string]bool
	pipModules map[string]bool
}

func (c *_DependencyCache) HasGoPackage(pkg string) bool {
	return c.goPackages[pkg]
}

func (c *_DependencyCache) AddGoPackage(pkg string) {
	if c.goPackages == nil {
		c.goPackages = make(map[string]bool)
	}
	c.goPackages[pkg] = true
}

func (c *_DependencyCache) HasPipModule(module string) bool {
	return c.pipModules[module]
}

func (c *_DependencyCache) AddPipModule(module string) {
	if c.pipModules == nil {
		c.pipModules = make(map[string]bool)
	}
	c.pipModules[module] = true
}

func (builder *Builder) buildTargetDependencies(target *repository.Target, targetOptions *_BuildTargetOption, options *_BuildTargetDependencyOptions) error {
	if target == nil {
		return errors.New("Require target")
	}

	if len(target.Spec().Dependencies) == 0 {
		return nil
	}

	log.Infof("Build dependencies of target [%v]", target.Name())
	log.Debugf("Builder.buildTargetDependencies: Build dependency of target [%v]. Options: %v", target.Name(), options.String())
	for _, depSpec := range target.Spec().Dependencies {
		if goDepSpec := depSpec.GetGo(); goDepSpec != nil && !options.GetNoGo() {
			// Build go dependency
			if err := builder.buildGoDependency(target, goDepSpec, targetOptions, options); err != nil {
				return err
			}
		} else if pipDepSpec := depSpec.GetPip(); pipDepSpec != nil && !options.GetNoPip() {
			// Build pip dependency
			if err := builder.buildPipDependency(target, pipDepSpec, targetOptions, options); err != nil {
				return err
			}
		} else if targetDepSpec := depSpec.GetTarget(); targetDepSpec != nil && !options.GetNoTarget() {
			// Build target dependency
			if err := builder.buildTargetDependency(target, targetDepSpec, targetOptions, options); err != nil {
				return err
			}
		} else {
			log.Warnf("No specific dependency specified in target [%v], skip", target.Name())
		}
	}

	// Done
	return nil
}

func (builder *Builder) buildGoDependency(sourceTarget *repository.Target, spec *pbSpec.GoDependency, targetOptions *_BuildTargetOption, options *_BuildTargetDependencyOptions) error {
	if spec.GetPackage() == "" {
		return errors.New("Require package")
	}

	log.Infof("Install go dependency [%v]", spec.Package)
	if builder.dependencyCache().HasGoPackage(spec.Package) {
		log.Infoln("Dependency already installed (cached)")
		return nil
	}

	// Run get get (-u) package
	args := []string{"get"}
	if options.GetUpdate() {
		args = append(args, "-u")
	}
	args = append(args, spec.Package)
	cmd := exec.Command("go", args...)
	if builder.Verbose() {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to install go package: %v", err)
		}
	} else {
		if outBytes, err := cmd.CombinedOutput(); err != nil {
			log.Errorln("Failed to install go package:\n", string(outBytes))
			return fmt.Errorf("Failed to install go package: %v", err)
		}
	}

	// Add cache and done
	builder.dependencyCache().AddGoPackage(spec.Package)
	return nil
}

func (builder *Builder) buildPipDependency(sourceTarget *repository.Target, spec *pbSpec.PipDependency, targetOptions *_BuildTargetOption, options *_BuildTargetDependencyOptions) error {
	if spec.GetModule() == "" {
		return errors.New("Require module")
	}

	log.Infof("Install pip dependency [%v]", spec.Module)
	if builder.dependencyCache().HasPipModule(spec.Module) {
		log.Infoln("Dependency already installed (cached)")
		return nil
	}

	// Run pip install (-U) module
	args := []string{"install"}
	if options.GetUpdate() {
		args = append(args, "-U")
	}
	args = append(args, spec.Module)
	cmd := exec.Command("pip", args...)
	if builder.Verbose() {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to install python module: %v", err)
		}
	} else {
		if outBytes, err := cmd.CombinedOutput(); err != nil {
			log.Errorln("Failed to install python module:\n", string(outBytes))
			return fmt.Errorf("Failed to install python module: %v", err)
		}
	}

	// Add cache and done
	builder.dependencyCache().AddPipModule(spec.Module)
	return nil
}

func (builder *Builder) buildTargetDependency(sourceTarget *repository.Target, spec *pbSpec.TargetDependency, targetOptions *_BuildTargetOption, options *_BuildTargetDependencyOptions) error {
	if sourceTarget == nil {
		return errors.New("Require source target")
	}
	if spec.GetTarget() == "" {
		return errors.New("Require target")
	}

	var target *repository.Target

	if spec.Reference != "" {
		// Get reference
		ref := sourceTarget.Package().GetReference(spec.Reference)
		if ref == nil {
			log.Errorf("Reference [%v] not found in package [%v]",
				spec.Reference,
				sourceTarget.Package().Spec().GetName(),
			)
		}
		// Get target
		var err error
		if target, err = builder.GetTargetByReference(ref, spec.Path, spec.Target); err != nil {
			return err
		}
		if target == nil {
			log.Errorf("Target [%v] with path[%v] in reference [%v] not found in package [%v]",
				spec.Target,
				spec.Path,
				spec.Reference,
				sourceTarget.Package().Spec().GetName(),
			)
			return fmt.Errorf("Target [%v] not found", spec.Target)
		}
	} else if spec.Path != "" {
		// The same repository
		var err error
		if target, err = sourceTarget.Package().GetRelativeTarget(spec.Path, spec.Target); err != nil {
			log.Errorf("Failed to get target [%v] with path [%v] in package [%v]: %v",
				spec.Target,
				spec.Path,
				sourceTarget.Package().Spec().GetName(),
				err,
			)
			return err
		}
		if target == nil {
			log.Errorf("Target [%v] with path[%v] not found in package [%v]",
				spec.Target,
				spec.Path,
				sourceTarget.Package().Spec().GetName(),
			)
			return fmt.Errorf("Target [%v] not found", spec.Target)
		}
	} else {
		// The same package
		target = sourceTarget.Package().GetTarget(spec.Target)
		if target == nil {
			log.Errorf("Target [%v] not found in package [%v]",
				spec.Target,
				sourceTarget.Package().Spec().GetName(),
			)
			return fmt.Errorf("Target [%v] not found", spec.Target)
		}
	}

	if result := builder.GetBuildResult(target); result != nil {
		log.Infoln("Dependency already built (cached)")
		return nil
	}

	// Build target
	newTargetOptions := *targetOptions
	newTargetOptions.noBuild = !options.BuildTarget || !spec.Build
	return builder.buildTarget(target, &newTargetOptions, options)
}

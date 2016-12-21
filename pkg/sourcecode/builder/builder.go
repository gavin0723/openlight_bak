// Author: lipixun
// Created Time : 日 10/23 21:26:33 2016
//
// File Name: build.go
// Description:
//
// 	The build process
//		1. Prepare stage:
//			a. Check the environment of the target in current builder. If not found, create a new one. Link the package to environment.
// 			b. Recursively visit the dependency to check the environment
// 			c. Until all packages are linked
// 		2. Build stage:
// 			a. Recursively build all targets with build spec defined, and collect the artifact
// 		3. [Optional] Copy stage:
// 			a. Copy the artifacts to output directory
//
// 	The environment struct
//		buildTempDir/
// 			environs/
// 				buildEnvironment/
//					...The linked packages, the structure depends on the build type...``
//					...The environment detail of each type will be documented at the header of source code file of each build type ...
// 			output/
//
//	The inject variables (all upper case)
//		BUILD_ENVIRON_[type]_PATH 			The environment (root) path for a specific build type
//		BUILD_TARGET_[target key]_PATH 		The target (root) path
//
//		When naming the variables, all chars except letters and underscore, will be replaced by underscore, and all letters will be converted to upper case
//

package builder

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode"
	"github.com/ops-openlight/openlight/pkg/sourcecode/graph"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"os"
	"path/filepath"
	"regexp"
)

const (
	BuilderLogHeader = "SourceCode.Builder"

	BuilderEnvironmentDirName = "environs"
	BuilderOutputDirName      = "output"

	BuilderDefaultArtifactName = "default"
)

var (
	TargetNameRegularExp = regexp.MustCompile("[^a-zA-Z\\d\\.]")

	SourceCodeBuilders map[string]SourceCodeBuilder = map[string]SourceCodeBuilder{
		BuilderTypeGolang: NewGolangSourceCodeBuilder(),
		BuilderTypePython: NewPythonSourceCodeBuilder(),
		BuilderTypeShell:  NewShellSourceCodeBuilder(),
		BuilderTypeDocker: NewDockerSourceCodeBuilder(),
	}
)

type Builder struct {
	graph           *graph.Graph
	logger          log.Logger
	path            string // The build temp path
	Options         BuilderOptions
	Results         map[string]*spec.BuildResult // The global build results, key is target key
	Environments    map[string]Environment       // The environments, key is build type
	preparedTargets map[string]bool              // The prepare targets
	builtTargets    map[string]bool              // The build targets
}

// Create a new Builder
func New(graph *graph.Graph, options BuilderOptions) (*Builder, error) {
	if graph == nil {
		return nil, errors.New("Require graph")
	}
	if options.Tag == "" {
		return nil, errors.New("Require tag")
	}
	// Get the build path
	path, err := graph.Workspace().Dir.User.GetPath(filepath.Join("sourcecode", "builder", options.Tag))
	if err != nil {
		return nil, err
	}
	// Create Builder
	return &Builder{
		graph:           graph,
		logger:          graph.Workspace().Logger.GetLoggerWithHeader(BuilderLogHeader),
		path:            path,
		Options:         options,
		Results:         make(map[string]*spec.BuildResult),
		Environments:    make(map[string]Environment),
		preparedTargets: make(map[string]bool),
		builtTargets:    make(map[string]bool),
	}, nil
}

// The graph of this builder
func (this *Builder) Graph() *graph.Graph {
	return this.graph
}

// The build path of this builder
func (this *Builder) Path() string {
	return this.path
}

func (this *Builder) EnvironmentPath() string {
	return filepath.Join(this.path, BuilderEnvironmentDirName)
}

func (this *Builder) OutputPath() string {
	return filepath.Join(this.path, BuilderOutputDirName)
}

// Build a target
func (this *Builder) Build(target *spec.Target) (*spec.BuildResult, error) {
	// Check if the target is in the graph
	if target == nil {
		return nil, errors.New("Require target")
	}
	// Check if has already built
	if result := this.Results[target.Key()]; result != nil {
		return result, nil
	}
	var err error
	// Stage 1. Prepare
	err = this.graph.Traverse(
		target,
		this.prepareGraphTraverseVisitor,
		nil,
		func(target *spec.Target, from *spec.Target, by *spec.TargetDependencySpec, action string, context interface{}) {
			ctx := context.(*BuilderContext)
			if action == graph.GraphTraverseActionEnter {
				ctx.Tracer.Push(sourcecode.TraceTypeTarget, target.Key(), target.Key())
			} else {
				ctx.Tracer.Pop()
			}
		},
		false,
		newBuilderContext(this),
	)
	if err != nil {
		return nil, err
	}
	// Stage 2. Build
	err = this.graph.Traverse(
		target,
		this.buildGraphTraverseVisitor,
		this.buildGraphTraverseController,
		func(target *spec.Target, from *spec.Target, by *spec.TargetDependencySpec, action string, context interface{}) {
			ctx := context.(*BuilderContext)
			if action == graph.GraphTraverseActionEnter {
				ctx.Tracer.Push(sourcecode.TraceTypeTarget, target.Key(), target.Key())
			} else {
				ctx.Tracer.Pop()
			}
		},
		false,
		newBuilderContext(this),
	)
	if err != nil {
		return nil, err
	}
	// Stage 3. Copy
	if this.Options.OutputPath != "" {
		if err := this.copy2Output(target); err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to copy artifact to output, error: %s", err))
		}
	}
	// Get the build result of the target and return
	result := this.Results[target.Key()]
	return result, nil
}

func (this *Builder) prepareGraphTraverseVisitor(target *spec.Target, from *spec.Target, by *spec.TargetDependencySpec, context interface{}) error {
	if !this.preparedTargets[target.Key()] {
		ctx := context.(*BuilderContext)
		this.logger.LeveledPrintf(log.LevelInfo, "Preparing %s\n", ctx.Tracer.String())
		builder := SourceCodeBuilders[target.Spec.Build.Type]
		if builder == nil {
			return errors.New(fmt.Sprintf("Builder [%s] not found", target.Spec.Build.Type))
		}
		// Get the environment
		environ, err := this.GetEnvironment(target.Spec.Build.Type)
		if err != nil {
			return err
		}
		// Prepare
		err = builder.Prepare(target, environ, ctx)
		if err != nil {
			return err
		}
		// Good, set prepared
		this.preparedTargets[target.Key()] = true
	}
	// Has already prepared
	return nil
}

func (this *Builder) buildGraphTraverseVisitor(target *spec.Target, from *spec.Target, by *spec.TargetDependencySpec, context interface{}) error {
	if !this.builtTargets[target.Key()] {
		ctx := context.(*BuilderContext)
		this.logger.LeveledPrintf(log.LevelInfo, "Building %s\n", ctx.Tracer.String())
		builder := SourceCodeBuilders[target.Spec.Build.Type]
		if builder == nil {
			return errors.New(fmt.Sprintf("Builder [%s] not found", target.Spec.Build.Type))
		}
		// Get the environment
		environ, err := this.GetEnvironment(target.Spec.Build.Type)
		if err != nil {
			return err
		}
		// Prepare
		err = builder.Build(target, environ, ctx)
		if err != nil {
			return err
		}
		// Good, set built
		this.builtTargets[target.Key()] = true
	}
	// Has already built
	return nil
}

func (this *Builder) buildGraphTraverseController(dep *spec.TargetDependencySpec, from *spec.Target, dest *spec.Target, context interface{}) bool {
	// Only build the dependency which is marked as build
	return dep.Options.Build
}

// Copy to output path
func (this *Builder) copy2Output(target *spec.Target) error {
	// Check output path
	if err := os.MkdirAll(this.Options.OutputPath, os.ModePerm); err != nil {
		return err
	}
	// Link to output
	buildResult := this.Results[target.Key()]
	if buildResult != nil {
		for _, art := range buildResult.Artifacts {
			if art.GetType() == artifact.ArtifactTypeFile {
				fileArtifact, ok := art.(*artifact.FileArtifact)
				if !ok {
					return errors.New("Cannot convert artifact to file artifact")
				}
				targetFile := filepath.Join(this.Options.OutputPath, target.Name, art.GetName())
				if fileArtifact.Compressed || fileArtifact.Files == nil {
					// The file artifact is a single file, add the file name
					targetFile = filepath.Join(targetFile, filepath.Base(fileArtifact.Path))
				}
				if _, err := os.Stat(targetFile); err == nil {
					// Remove it
					if err := os.Remove(targetFile); err != nil {
						return err
					}
				} else if !os.IsNotExist(err) {
					return err
				}
				// Ensure the target file directory
				targetDir := filepath.Dir(targetFile)
				if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
					return err
				}
				// Link
				if err := os.Symlink(fileArtifact.Path, targetFile); err != nil {
					return err
				}
			}
		}
	}
	// Done
	return nil
}

// Get the environment of build type t
func (this *Builder) GetEnvironment(t string) (Environment, error) {
	environ := this.Environments[t]
	if environ == nil {
		builder := SourceCodeBuilders[t]
		if builder == nil {
			return nil, errors.New(fmt.Sprintf("Unknown type [%s]", t))
		}
		environ, err := builder.NewEnviron(this)
		if err != nil {
			return nil, err
		}
		this.Environments[t] = environ
		return environ, nil
	} else {
		return environ, nil
	}
}

// Ensure the target output path
func (this *Builder) EnsureTargetOutputPath(target *spec.Target) (string, error) {
	path := filepath.Join(this.OutputPath(), GetTargetRegularKey(target))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}

func (this *Builder) NewBuildMetadata(target *spec.Target) spec.BuildMetadata {
	return spec.BuildMetadata{
		Tag:        this.Options.Tag,
		Time:       this.Options.Time,
		Repository: target.Repository.Metadata,
		SourcePath: target.Path(),
	}
}

func (this *Builder) SetBuildResultDependency(target *spec.Target, buildResult *spec.BuildResult) {
	for name, dep := range target.Spec.Deps {
		depBuildResult := this.Results[dep.Key()]
		if depBuildResult != nil {
			buildResult.Deps[name] = depBuildResult
		}
	}
}

func (this *Builder) AddResult(target *spec.Target, buildResult *spec.BuildResult) {
	this.Results[target.Key()] = buildResult
}

// Get the target regular key
func GetTargetRegularKey(target *spec.Target) string {
	return TargetNameRegularExp.ReplaceAllString(target.Key(), "_")
}

type BuilderContext struct {
	Graph     *graph.Graph         // The graph
	Builder   *Builder             // The current builder
	Tracer    *sourcecode.Tracer   // The build tracer
	Workspace *workspace.Workspace // The workspace
}

// Create a new BuilderContext
func newBuilderContext(builder *Builder) *BuilderContext {
	return &BuilderContext{
		Graph:     builder.Graph(),
		Builder:   builder,
		Tracer:    sourcecode.NewTracer(),
		Workspace: builder.Graph().Workspace(),
	}
}

type SourceCodeBuilder interface {
	// Create new environment for the builder
	NewEnviron(builder *Builder) (Environment, error)
	// Prepare for the target
	Prepare(target *spec.Target, env Environment, context *BuilderContext) error
	// Build the target
	Build(target *spec.Target, env Environment, context *BuilderContext) error
}

// Clean all build data
func CleanBuildData(ws *workspace.Workspace) error {
	path, err := ws.Dir.User.GetPath(filepath.Join("sourcecode", "builder"))
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

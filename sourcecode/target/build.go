// Author: lipixun
// Created Time : 日 10/23 21:26:33 2016
//
// File Name: build.go
// Description:
//
package target

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/sourcecode/spec"
	"github.com/ops-openlight/openlight/sourcecode/workspace"
	"os"
	"path/filepath"
	"time"
)

// The builder interface
type Builder interface {
	Type() string
	Build(ctx *TargetBuildContext) (*BuildResult, error)
}

// The build option
type BuildOption struct {
	Tag         string
	Time        time.Time
	Copy2Output bool
}

// Create a new BuildOption
func NewBuildOption(tag string) *BuildOption {
	return &BuildOption{
		Tag:         tag,
		Copy2Output: true,
		Time:        time.Now(),
	}
}

type TargetGraphBuilder struct {
	Option    *BuildOption
	Graph     *TargetGraph
	Workspace *workspace.Workspace
	Globals   map[string]*BuildResult // The global build results, key is target key
}

// Create a new TargetGraphBuilder
func NewTargetGraphBuilder(option *BuildOption, graph *TargetGraph, ws *workspace.Workspace) (*TargetGraphBuilder, error) {
	if option == nil {
		return nil, errors.New("Require option")
	}
	if graph == nil {
		return nil, errors.New("Require graph")
	}
	if ws == nil {
		return nil, errors.New("Require workspace")
	}
	// Create TargetGraphBuilder
	return &TargetGraphBuilder{
		Option:    option,
		Graph:     graph,
		Workspace: ws,
		Globals:   make(map[string]*BuildResult),
	}, nil
}

// Build a target
func (this *TargetGraphBuilder) Build(target *Target) (*BuildResult, error) {
	// Check if the target is in the graph
	if target == nil {
		return nil, errors.New("Require target")
	}
	if !this.Graph.HasTarget(target.Key()) {
		return nil, errors.New("Target not found in this graph")
	}
	// Check the global result
	if r, ok := this.Globals[target.Key()]; ok {
		return r, nil
	}
	// Create the build context and build
	ctx := NewTargetBuildContext(target, this.Option, this.Graph, this, this.Workspace)
	result, err := ctx.Build()
	if err != nil {
		return nil, err
	}
	// Check copy to output
	if this.Option.Copy2Output {
		// Link the gen output path to output path
		outputPath, err := this.Workspace.FileSystem.GetOutputPath(true)
		if err != nil {
			return nil, errors.New(fmt.Sprint("Failed to get output path, error: ", err))
		}
		outputPath = filepath.Join(outputPath, target.Repository.Uri(), target.Name)
		if _, err := os.Stat(outputPath); err != nil {
			if !os.IsNotExist(err) {
				// Unexpected error
				return nil, errors.New(fmt.Sprint("Failed to check output path, error: ", err))
			}
		} else {
			// Remove it
			if err := os.Remove(outputPath); err != nil {
				return nil, errors.New(fmt.Sprint("Failed to remove previous build result, error: ", err))
			}
		}
		if result.OutputPath != "" {
			if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to prepare output path dir [%s], error: %s", filepath.Dir(outputPath), err))
			}
			if err := os.Symlink(result.OutputPath, outputPath); err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to link build result, error: %s", err))
			}
		}
	}
	// Done
	return result, nil
}

type TargetBuildContext struct {
	Option    *BuildOption
	Target    *Target
	Graph     *TargetGraph
	Builder   *TargetGraphBuilder
	Path      *TargetDerivationPath // The path which shows the build dependency
	Workspace *workspace.Workspace
	completed bool
}

// Create a new TargetBuildContext
func NewTargetBuildContext(target *Target, option *BuildOption, graph *TargetGraph, builder *TargetGraphBuilder, ws *workspace.Workspace) *TargetBuildContext {
	// Push this target into path
	buildPath := NewTargetDerivationPath()
	buildPath.Push(NewTargetDerivationFromTarget(target))
	// Done
	return &TargetBuildContext{
		Graph:     graph,
		Target:    target,
		Builder:   builder,
		Option:    option,
		Workspace: ws,
		Path:      buildPath,
		completed: false,
	}
}

func (this *TargetBuildContext) IsBuilding(key string) bool {
	// Check if a target is building
	return this.Path.HasKey(key)
}

// Build the target
func (this *TargetBuildContext) Build() (*BuildResult, error) {
	// Check if built
	if result, ok := this.Builder.Globals[this.Target.Key()]; ok {
		return result, nil
	}
	// Check if completed
	if this.completed {
		return nil, errors.New("Build context has already completed")
	}
	// Check if we have builder here
	if this.Target.Builder == nil {
		// Log builder defined, warning
		this.Workspace.Logger.WriteWarningHeaderln(TargetGraphLogHeader, "Target [", this.Target.Name, "] builder not found")
		// Mark as completed
		if err := this.complete(nil); err != nil {
			return nil, err
		}
		// We donnot return an error here
		return nil, nil
	}
	// Build the target
	this.Workspace.Logger.WriteInfoHeaderln(TargetGraphLogHeader, "Build ", this.PathString())
	result, err := this.Target.Builder.Build(this)
	if err != nil {
		return nil, err
	}
	// Mark as completed
	if err := this.complete(result); err != nil {
		return nil, err
	}
	// Done
	return result, nil
}

// Derive a build context
func (this *TargetBuildContext) Derive(dep *TargetDependency) (*TargetBuildContext, error) {
	// Check if current target is building
	if this.IsBuilding(dep.Target.Key()) {
		// A circle build dependency
		return nil, errors.New(fmt.Sprintf("Circle building chain found, conflict target [%s], building chain: %s", dep.Target.Name, this.PathString()))
	}
	// OK, push path
	this.Path.Push(NewTargetDerivationFromDep(dep))
	// Create new context
	return &TargetBuildContext{
		Target:    dep.Target,
		Graph:     this.Graph,
		Option:    this.Option,
		Builder:   this.Builder,
		Path:      this.Path,
		Workspace: this.Workspace,
	}, nil
}

// Complete this context
func (this *TargetBuildContext) complete(result *BuildResult) error {
	lastDeriveTarget := this.Path.Last()
	if lastDeriveTarget == nil {
		return errors.New("No more target building to complete")
	}
	if lastDeriveTarget.Target != this.Target {
		return errors.New(fmt.Sprintf(
			"Inconsistent target building to complete. Expect [%s] Actual [%s]",
			this.Target.Key(),
			lastDeriveTarget.Target.Key(),
		))
	}
	if _, ok := this.Builder.Globals[this.Target.Key()]; ok {
		return errors.New("Target building has already completed")
	}
	// Complete
	if result != nil {
		this.Builder.Globals[this.Target.Key()] = result
	}
	this.Path.Pop()
	// Done
	return nil
}

// Get the path string
func (this *TargetBuildContext) PathString() string {
	return this.Path.String()
}

type BuildResult struct {
	Metadata   BuildMetadata      `json:"metadata"`
	OutputPath string             `json:"outputPath"`
	Artifacts  []*Artifact        `json:"artifacts"`
	Deps       []*BuildDependency `json:"deps"`
}

type BuildMetadata struct {
	Tag  string    `json:"tag"`
	Time time.Time `json:"time"`
}

type BuildDependency struct {
	Spec   spec.TargetDependency `json:"spec" yaml:"spec"`
	Result *BuildResult          `json:"result" yaml:"result"` // The build result of the dependent target
}

// Create a new TargetGraphBuilder
func (this *TargetGraph) NewBuilder(option *BuildOption, ws *workspace.Workspace) (*TargetGraphBuilder, error) {
	return NewTargetGraphBuilder(option, this, ws)
}

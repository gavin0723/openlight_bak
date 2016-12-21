// Author: lipixun
// Created Time : 四 10/20 18:39:54 2016
//
// File Name: graph.go
// Description:
//	Target graph
package graph

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode"
	"github.com/ops-openlight/openlight/pkg/sourcecode/repofinder"
	"github.com/ops-openlight/openlight/pkg/sourcecode/repoloader"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"strings"
)

const (
	GraphLogHeader          = "SourceCode.Graph"
	RepositoryDefaultBranch = "master"

	GraphTraverseActionEnter = "enter"
	GraphTraverseActionLeave = "leave"
)

type Graph struct {
	ws               *workspace.Workspace
	logger           log.Logger
	Options          GraphOptions
	Repositories     map[string]*spec.Repository
	Targets          map[string]*spec.Target
	RemoteOverwrites map[string]string // Key is uri, value is remote
}

type GraphOptions struct {
	UseLocalDependency bool // Whether to use local repository to resolve the dependency
	DisableFinder      bool
}

func New(ws *workspace.Workspace, options GraphOptions) (*Graph, error) {
	if ws == nil {
		return nil, errors.New("Require workspace")
	}
	// Create a new target graph
	return &Graph{
		ws:               ws,
		logger:           ws.Logger.GetLogger(ws.Logger.GetLevel(), ws.Logger.GetDefaultLevel(), GraphLogHeader),
		Options:          options,
		Repositories:     make(map[string]*spec.Repository),
		Targets:          make(map[string]*spec.Target),
		RemoteOverwrites: make(map[string]string),
	}, nil
}

func (this *Graph) Workspace() *workspace.Workspace {
	return this.ws
}

type LoadOptions struct {
	Uri     string // The expected uri of the loading repository
	Type    string
	Branch  string
	Commit  string
	Targets []string
}

// Load a repository
func (this *Graph) Load(remote string, options LoadOptions) (*spec.Repository, error) {
	return this.load(remote, options, sourcecode.NewTracer())
}

// Load a repository
func (this *Graph) load(remote string, options LoadOptions, tracer *sourcecode.Tracer) (*spec.Repository, error) {
	this.logger.LeveledPrintf(log.LevelDebug, "Load repository: %s\n", remote)
	// Push into tracer
	if options.Uri != "" {
		tracer.Push(sourcecode.TraceTypeRepository, options.Uri, options.Uri)
	} else {
		tracer.Push(sourcecode.TraceTypeRepository, "<root>", "<root>")
	}
	defer tracer.Pop()
	this.logger.LeveledPrintf(log.LevelInfo, "Loading %s\n", tracer.String())
	// Rewrite the remote
	if options.Uri != "" {
		if _remote, ok := this.RemoteOverwrites[options.Uri]; ok {
			this.logger.LeveledPrintf(log.LevelDebug, "Overwrite remote of repository [%s] from [%s] to [%s]\n", options.Uri, remote, _remote)
			remote = _remote
		}
	}
	// Check the loaded repositories
	if options.Uri != "" {
		loadedRepo, ok := this.Repositories[options.Uri]
		if ok {
			// Compare the repository source
			if loadedRepo.Source != remote {
				this.logger.LeveledPrintf(log.LevelError, "Conflict source of repository [%s]. Loaded [%s] Requested [%s]\n", options.Uri, loadedRepo.Uri, remote)
				return nil, errors.New("Conflict repository source")
			}
			// Resolve this repository
			if err := this.resolve(loadedRepo, options.Targets, tracer); err != nil {
				return nil, err
			}
			// Done
			return loadedRepo, nil
		}
	}
	// Load this repository
	t := options.Type
	if t == "" {
		t = spec.DefaultRepositoryType
	}
	loader := repoloader.GetLoader(t)
	if loader == nil {
		return nil, errors.New(fmt.Sprintf("Repository loader for type [%s] not found", t))
	}
	loadingRepo, err := loader.Load(remote, repoloader.LoadOptions{Branch: options.Branch, Commit: options.Commit}, this.ws)
	if err != nil {
		return nil, err
	}
	if options.Uri != "" && loadingRepo.Uri != options.Uri {
		this.logger.LeveledPrintf(log.LevelError, "Mismatch repository uri. Expected [%s] Actually [%s]\n", options.Uri, loadingRepo.Uri)
		return nil, errors.New("Mismatch repository uri")
	}
	loadedRepo, ok := this.Repositories[loadingRepo.Uri]
	if ok {
		// Compare the two repository
		if loadingRepo.Source != loadedRepo.Source {
			this.logger.LeveledPrintf(log.LevelError, "Conflict source of repository [%s]. Loaded [%s] Requested [%s]\n", options.Uri, loadedRepo.Uri, remote)
			return nil, errors.New("Conflict repository source")
		}
		// Use the loaded repository, resolve it
		if err := this.resolve(loadedRepo, options.Targets, tracer); err != nil {
			return nil, err
		}
		// Done
		return loadedRepo, nil
	} else {
		// Add this repository
		this.Repositories[loadingRepo.Uri] = loadingRepo
		// Resolve this repository
		if err := this.resolve(loadingRepo, options.Targets, tracer); err != nil {
			return nil, err
		}
		// Done
		return loadingRepo, nil
	}
}

// Resolve a repository
func (this *Graph) resolve(r *spec.Repository, targets []string, tracer *sourcecode.Tracer) error {
	if len(targets) == 0 {
		// Load all targets in this repository
		for targetName, targetSpec := range r.Spec.Targets {
			_, err := this.loadTarget(targetName, targetSpec, r, tracer)
			if err != nil {
				return err
			}
		}
	} else {
		for _, targetName := range targets {
			targetSpec, ok := r.Spec.Targets[targetName]
			if !ok {
				this.logger.LeveledPrintf(log.LevelError, "Target [%s] not found in repository [%s]\n", targetName, r.Uri)
				return errors.New(fmt.Sprintf("Target [%s] not found in repository [%s]", targetName, r.Uri))
			}
			_, err := this.loadTarget(targetName, targetSpec, r, tracer)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Graph) loadTarget(targetName string, targetSpec *spec.TargetSpec, r *spec.Repository, tracer *sourcecode.Tracer) (*spec.Target, error) {
	this.logger.LeveledPrintf(log.LevelDebug, "Load target [%s] from repository [%s]\n", targetName, r.Uri)
	targetKey := spec.GetTargetKey(targetName, r)
	// Check loop
	if tracer.Has(sourcecode.TraceTypeTarget, targetKey) {
		this.logger.LeveledPrintf(log.LevelError, "Target loading loop found on target [%s] from repository [%s]. Loading trace path: %s\n", targetName, r.Uri, tracer.String())
		return nil, errors.New("Target loading loop found")
	}
	// Load the target
	target, ok := this.Targets[targetKey]
	if ok {
		// Loaded
		return target, nil
	}
	// Push into tracer
	tracer.Push(sourcecode.TraceTypeTarget, targetKey, targetName)
	defer tracer.Pop()
	this.logger.LeveledPrintf(log.LevelInfo, "Loading %s\n", tracer.String())
	target = &spec.Target{
		Name:       targetName,
		Repository: r,
		Spec:       targetSpec,
	}
	// Resolve the dependency
	for depName, depSpec := range target.Spec.Deps {
		if depSpec.Repository == "" {
			// Set the repository to the repository of current target
			depSpec.Repository = r.Uri
		}
		_, ok := this.Targets[depSpec.Key()]
		if ok {
			continue
		}
		// Resolve this dependency
		if err := this.resolveTargetDependency(depName, depSpec.Target, depSpec.Repository, target, tracer); err != nil {
			return nil, err
		}
	}
	// Good, add this target
	this.Targets[targetKey] = target
	return target, nil
}

func (this *Graph) resolveTargetDependency(name, targetName, repository string, target *spec.Target, tracer *sourcecode.Tracer) error {
	this.logger.LeveledPrintf(log.LevelDebug, "Resolve target dependency [%s] from repository [%s] target [%s]\n", name, repository, targetName)
	// Push tracer
	tracer.Push(sourcecode.TraceTypeDependency, spec.GetTargetDependencyKey(name, targetName, repository), name)
	defer tracer.Pop()
	this.logger.LeveledPrintf(log.LevelInfo, "Loading %s\n", tracer.String())
	// Load the repository and target
	if repository == target.Repository.Uri {
		// Load the target from repository itself
		targetSpec, ok := target.Repository.Spec.Targets[targetName]
		if !ok {
			this.logger.LeveledPrintf(log.LevelError, "Target [%s] not found in repository [%s]\n", targetName, target.Repository.Uri)
			return errors.New("Dependent target not found")
		}
		_, err := this.loadTarget(targetName, targetSpec, target.Repository, tracer)
		return err
	}
	// Get the repository reference info
	refer, ok := target.Repository.Spec.References[repository]
	if !ok {
		this.logger.LeveledPrintf(log.LevelError, "Repository reference of [%s] not found\n", repository)
		return errors.New("Repository reference not found")
	}
	remote := refer.Remote
	// Check the local
	if this.Options.UseLocalDependency && !this.Options.DisableFinder && refer.Finder.Type != "" {
		// Find the repository by finder
		finder := repofinder.GetFinder(refer.Finder.Type)
		if finder == nil {
			this.logger.LeveledPrintf(log.LevelWarn, "Repository finder for [%s] with type [%s] not found\n", repository, refer.Finder.Type)
		} else {
			paths, err := finder.Find(this.ws, refer.Finder.Params)
			if err != nil {
				this.logger.LeveledPrintf(log.LevelWarn, "Failed to find repository for [%s] with type [%s] error: %s\n", repository, refer.Finder.Type, err)
			} else if len(paths) > 2 {
				this.logger.LeveledPrintf(log.LevelWarn, "Too many repository found by finder [%s] for repository [%s], found: %s\n", refer.Finder.Type, repository, strings.Join(paths, ", "))
			} else if len(paths) == 1 {
				this.logger.LeveledPrintf(log.LevelWarn, "Found repository for [%s] with type [%s]: %s\n", repository, refer.Finder.Type, paths[0])
				remote = paths[0]
			} else {
				this.logger.LeveledPrintf(log.LevelDebug, "No repository found for [%s] with type [%s]\n", repository, refer.Finder.Type)
			}
		}
	}
	// Load it
	_, err := this.load(remote, LoadOptions{Uri: repository, Branch: refer.Branch, Commit: refer.Commit}, tracer)
	return err
}

// The visitor to traverse the graph
type TraverseNotifier func(target *spec.Target, from *spec.Target, by *spec.TargetDependencySpec, action string, context interface{})
type TraverseVisitor func(target *spec.Target, from *spec.Target, by *spec.TargetDependencySpec, context interface{}) error
type TraverseController func(dep *spec.TargetDependencySpec, from *spec.Target, dest *spec.Target, context interface{}) bool

// Traverse the graph from a target
func (this *Graph) Traverse(target *spec.Target, visitor TraverseVisitor, controller TraverseController, notifier TraverseNotifier, preorder bool, context interface{}) error {
	if visitor == nil {
		return errors.New("Require visitor")
	}
	if _, ok := this.Targets[target.Key()]; !ok {
		return errors.New("Target not found in current graph")
	}
	return this._traverse(target, nil, nil, visitor, controller, notifier, preorder, context)
}

func (this *Graph) _traverse(
	target *spec.Target,
	from *spec.Target,
	by *spec.TargetDependencySpec,
	visitor TraverseVisitor,
	controller TraverseController,
	notifier TraverseNotifier,
	preorder bool,
	context interface{},
) error {
	if notifier != nil {
		notifier(target, from, by, GraphTraverseActionEnter, context)
	}
	if preorder {
		if err := visitor(target, from, by, context); err != nil {
			return err
		}
	}
	// Visit all dependency
	for _, dep := range target.Spec.Deps {
		depTarget := this.Targets[dep.Key()]
		if depTarget == nil {
			return errors.New(fmt.Sprintf("Dependency target [%s] not found", dep.Key()))
		}
		if controller == nil || controller(dep, target, depTarget, context) {
			// Traverse
			if err := this._traverse(depTarget, target, dep, visitor, controller, notifier, preorder, context); err != nil {
				return err
			}
		}
	}
	if !preorder {
		if err := visitor(target, from, by, context); err != nil {
			return err
		}
	}
	if notifier != nil {
		notifier(target, from, by, GraphTraverseActionLeave, context)
	}
	// Done
	return nil
}

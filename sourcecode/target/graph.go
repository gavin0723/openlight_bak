// Author: lipixun
// Created Time : 四 10/20 18:39:54 2016
//
// File Name: graph.go
// Description:
//	Target graph
package target

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/sourcecode/repository"
	"github.com/ops-openlight/openlight/sourcecode/workspace"
	"github.com/ops-openlight/openlight/uri"
	"strings"
)

const (
	TargetGraphLogHeader = "TargetGraph"
)

type TargetGraph struct {
	targets map[string]*Target
}

func NewTargetGraph() *TargetGraph {
	// Create a new target graph
	return &TargetGraph{
		targets: make(map[string]*Target),
	}
}

func (this *TargetGraph) HasTarget(key string) bool {
	_, ok := this.targets[key]
	return ok
}

func (this *TargetGraph) Load(target string, repo *repository.Repository, ws *workspace.Workspace) (*Target, error) {
	// Load the graph from a target
	if repo == nil {
		return nil, errors.New("Require repository")
	}
	// Load the target
	return this.loadTarget(target, repo, ws, NewTargetDerivationPath())
}

func (this *TargetGraph) loadTarget(target string, repo *repository.Repository, ws *workspace.Workspace, derivePath *TargetDerivationPath) (*Target, error) {
	// Load the target
	ws.Logger.WriteInfoHeaderln(TargetGraphLogHeader, "Load ", fmt.Sprintf("%s:%s", repo.Uri(), target), " from ", derivePath.String())
	if target == "" {
		// Try to use the default target
		target = repo.Spec.Options.Default.Target
		if target == "" {
			return nil, errors.New(fmt.Sprintf("Default target not found in repository [%s]", repo.Uri()))
		}
	}
	key := GetTargetKey(target, repo)
	if t, ok := this.targets[key]; ok {
		return t, nil
	}
	// Find the target in the repository
	targetSpec, _ := repo.Spec.Targets[target]
	if targetSpec == nil {
		// Target not found
		return nil, errors.New(fmt.Sprintf("Target [%s] not found in repository [%s]", target, repo.Uri()))
	}
	// Load this target
	t, err := NewTarget(target, targetSpec, repo)
	if err != nil {
		return nil, err
	}
	// NOTE: Here we add this target before solving deps, this will let circle dependency work
	this.targets[t.Key()] = t
	// Resolve deps
	for _, depSpec := range targetSpec.Deps {
		dep := &TargetDependency{Spec: depSpec}
		// Push this dep
		derivePath.Push(NewTargetDerivationFromDep(dep))
		// Load dep target
		if dep.Spec.Repository != nil {
			// Load the repository
			depRepo, err := ws.Repository.Load(dep.Spec.Repository.Uri, &uri.RepositoryReference{Remote: dep.Spec.Repository.Remote}, ws.DefaultRepositoryPathFunction)
			if err != nil {
				return nil, err
			}
			if ws.Verbose() {
				ws.Logger.WriteDebugHeaderln(TargetGraphLogHeader, "Loaded repository [", depRepo.Uri(), "] with metadata: ", depRepo.Metadata.String())
			}
			dep.Target, err = this.loadTarget(dep.Spec.Target, depRepo, ws, derivePath)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			dep.Target, err = this.loadTarget(dep.Spec.Target, repo, ws, derivePath)
			if err != nil {
				return nil, err
			}
		}
		// Pop
		derivePath.Pop()
		// Set the dep
		t.Deps = append(t.Deps, dep)
	}
	// Done
	return t, nil
}

type TargetDerivation struct {
	Dep    *TargetDependency
	Target *Target
}

func NewTargetDerivationFromTarget(target *Target) *TargetDerivation {
	return &TargetDerivation{Target: target}
}

func NewTargetDerivationFromDep(dep *TargetDependency) *TargetDerivation {
	return &TargetDerivation{Dep: dep, Target: dep.Target}
}

type TargetDerivationPath struct {
	derivations []*TargetDerivation
	keys        map[string]int
}

func NewTargetDerivationPath() *TargetDerivationPath {
	return &TargetDerivationPath{
		keys: make(map[string]int),
	}
}

func (this *TargetDerivationPath) HasKey(key string) bool {
	// Check if a key in this path
	count, ok := this.keys[key]
	return ok && count > 0
}

func (this *TargetDerivationPath) Last() *TargetDerivation {
	if len(this.derivations) > 0 {
		return this.derivations[len(this.derivations)-1]
	} else {
		return nil
	}
}

func (this *TargetDerivationPath) Push(derivation *TargetDerivation) {
	// Push a derivation
	this.derivations = append(this.derivations, derivation)
	// Check key
	if derivation.Target != nil {
		key := derivation.Target.Key()
		if count, ok := this.keys[key]; ok {
			this.keys[key] = count + 1
		} else {
			this.keys[key] = 1
		}
	}
}

func (this *TargetDerivationPath) Pop() *TargetDerivation {
	// Pop a derive
	if len(this.derivations) > 0 {
		// Pop up the last one
		derivation := this.derivations[len(this.derivations)-1]
		this.derivations = this.derivations[:len(this.derivations)-1]
		if derivation.Target != nil {
			key := derivation.Target.Key()
			if count, ok := this.keys[key]; ok {
				if count > 1 {
					this.keys[key] = count - 1
				} else {
					delete(this.keys, key)
				}
			}
		}
		return derivation
	} else {
		return nil
	}
}

func (this *TargetDerivationPath) String() string {
	var strs []string
	for _, derive := range this.derivations {
		if derive.Target != nil {
			strs = append(strs, derive.Target.Key())
		} else if derive.Dep != nil {
			if derive.Dep.Spec.Repository != nil {
				strs = append(strs, fmt.Sprintf("%s:%s", derive.Dep.Spec.Repository.Uri, derive.Dep.Spec.Target))
			} else {
				strs = append(strs, derive.Dep.Spec.Target)
			}
		} else {
			strs = append(strs, "?")
		}
	}
	return strings.Join(strs, " ==> ")
}

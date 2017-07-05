// Author: lipixun
// File Name: context.go
// Description:

package builder

import (
	"github.com/ops-openlight/openlight/pkg/builder/buildcontext"
	"github.com/ops-openlight/openlight/pkg/repository"
)

// Ensure the interface is implemented
var _ buildcontext.Context = (*BuildContext)(nil)

// BuildContext defines the build context
type BuildContext struct {
	verbose         bool
	tag             string
	depCache        _DependencyCache
	rootRepo        *repository.LocalRepository
	remoteRepos     map[string]*repository.LocalRepository
	buildingTargets map[string]bool
	buildResults    map[string]*buildcontext.TargetBuildResult
}

func (ctx *BuildContext) dependencyCache() *_DependencyCache {
	return &ctx.depCache
}

// Verbose returns verbose or not
func (ctx *BuildContext) Verbose() bool {
	return ctx.verbose
}

// Tag returns the build tag
func (ctx *BuildContext) Tag() string {
	return ctx.tag
}

// RootRepository returns the root repository of this context
func (ctx *BuildContext) RootRepository() *repository.LocalRepository {
	return ctx.rootRepo
}

// GetRemoteRepository returns the loaded repository with a remote uri
func (ctx *BuildContext) GetRemoteRepository(remote string) *repository.LocalRepository {
	return ctx.remoteRepos[remote]
}

// SetRemoteRepository sets a local repository with a remote uri
func (ctx *BuildContext) SetRemoteRepository(remote string, repo *repository.LocalRepository) {
	if ctx.remoteRepos == nil {
		ctx.remoteRepos = make(map[string]*repository.LocalRepository)
	}
	ctx.remoteRepos[remote] = repo
}

// IsTargetBuilding checks if the target is building or not
func (ctx *BuildContext) IsTargetBuilding(target *repository.Target) bool {
	return ctx.buildingTargets[target.Key()]
}

func (ctx *BuildContext) setTargetBuilding(target *repository.Target) {
	if ctx.buildingTargets == nil {
		ctx.buildingTargets = make(map[string]bool)
	}
	ctx.buildingTargets[target.Key()] = true
}

func (ctx *BuildContext) unsetTargetBuilding(target *repository.Target) {
	if ctx.buildingTargets != nil {
		delete(ctx.buildingTargets, target.Key())
	}
}

// GetBuildResult returns the build result of a target
func (ctx *BuildContext) GetBuildResult(target *repository.Target) *buildcontext.TargetBuildResult {
	return ctx.buildResults[target.Key()]
}

// GetBuildResults returns all build resutls
func (ctx *BuildContext) GetBuildResults() []*buildcontext.TargetBuildResult {
	var results []*buildcontext.TargetBuildResult
	for _, r := range ctx.buildResults {
		results = append(results, r)
	}
	return results
}

func (ctx *BuildContext) addBuildResult(target *repository.Target, result *buildcontext.TargetBuildResult) {
	if ctx.buildResults == nil {
		ctx.buildResults = make(map[string]*buildcontext.TargetBuildResult)
	}
	ctx.buildResults[target.Key()] = result
}

// GetTargetOutputDir prepares and returns the target output dir
func (ctx *BuildContext) GetTargetOutputDir(target *repository.Target, init bool) (string, error) {
	if init {
		return target.InitOutputDir("build")
	}
	return target.GetOutputDir("build"), nil
}

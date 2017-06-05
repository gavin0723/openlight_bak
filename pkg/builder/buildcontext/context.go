// Author: lipixun
// File Name: context.go
// Description:

package buildcontext

import (
	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/repository"
)

// Context defines the build context
type Context interface {
	// Verbose returns verbose or not
	Verbose() bool
	// Tag returns the build tag
	Tag() string
	// GetBuildResult returns the build result of a target
	GetBuildResult(target *repository.Target) *TargetBuildResult
	// GetTargetOutputDir prepares and returns the target output dir
	GetTargetOutputDir(target *repository.Target, init bool) (string, error)
	// GetRemoteRepository returns the loaded repository with a remote uri
	GetRemoteRepository(remote string) *repository.LocalRepository
}

// TargetBuildResult defines the target build result
type TargetBuildResult struct {
	artifact artifact.Artifact
}

// NewTargetBuildResult creates a new TargetBuildResult
func NewTargetBuildResult(art artifact.Artifact) *TargetBuildResult {
	return &TargetBuildResult{art}
}

// Artifact returns the artifact
func (result *TargetBuildResult) Artifact() artifact.Artifact {
	return result.artifact
}

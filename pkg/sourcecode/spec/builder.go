// Author: lipixun
// Created Time : 一 12/12 22:24:17 2016
//
// File Name: builder.go
// Description:
//	The builder spec
package spec

import (
	"github.com/ops-openlight/openlight/pkg/artifact"
	"time"
)

type SourceCodeLink struct {
	Path       string `yaml:"path"`       // The relative path of the linked source in the repository
	LinkedName string `yaml:"linkedName"` // The name of the linked target, will use the original name if not specified
}

// The build result of a target
type BuildResult struct {
	Repository string                       `json:"repository"` // The repository uri
	Target     string                       `json:"target"`     // The target name
	Metadata   BuildMetadata                `json:"metadata"`   // The metadata
	Artifacts  map[string]artifact.Artifact `json:"artifacts"`  // All collected artifacts
	Deps       map[string]*BuildResult      `json:"deps"`       // The build results of dependencies, name is the dep name
}

type BuildMetadata struct {
	Tag            string                 `json:"tag"`            // The build tag
	Time           time.Time              `json:"time"`           // The time when start the build
	Builder        string                 `json:"builder"`        // The builder type
	BuildTimeUsage float64                `json:"buildTimeUsage"` // The build time in seconds
	BuildParams    map[string]interface{} `json:"buildParams"`    // The build parameters
	Repository     RepositoryMetadata     `json:"repository"`     // The repository metadata
	SourcePath     string                 `json:"sourcePath"`     // The source path (root source path)
	LinkedPath     string                 `json:"linkedPath"`     // The linked path in the build environment (root linked path)
	OutputPath     string                 `json:"outputPath"`     // The build output path (root output path)
}

func NewBuildResult(target *Target, metadata BuildMetadata) *BuildResult {
	return &BuildResult{
		Repository: target.Repository.Uri,
		Target:     target.Name,
		Metadata:   metadata,
		Artifacts:  make(map[string]artifact.Artifact),
		Deps:       make(map[string]*BuildResult),
	}
}

// Author: lipixun
// Created Time : 一 12/19 14:51:32 2016
//
// File Name: artifact.go
// Description:
//	The generate artifact collector
package builder

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"path/filepath"
	"regexp"
)

func CollectFileArtifactBySpecs(path string, specs map[string]*spec.FileArtifactCollectorSpec) ([]artifact.Artifact, error) {
	var arts []artifact.Artifact
	for name, artSpec := range specs {
		art, err := CollectFileArtifactBySpec(name, filepath.Join(path, artSpec.Path), artSpec)
		if err != nil {
			return nil, err
		}
		arts = append(arts, art)
	}
	return arts, nil
}

func CollectFileArtifactBySpec(name, path string, artSpec *spec.FileArtifactCollectorSpec) (artifact.Artifact, error) {
	options := artifact.NewDefaultCollectFileArtifactOptions()
	if artSpec.Includes != "" {
		exp, err := regexp.Compile(artSpec.Includes)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to compile artifact includes regular expression [%s], error: %s", artSpec.Includes, err))
		}
		options.Includes = exp
	}
	if artSpec.Excludes != "" {
		exp, err := regexp.Compile(artSpec.Excludes)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to compile artifact excludes regular expression [%s], error: %s", artSpec.Excludes, err))
		}
		options.Excludes = exp
	}
	options.Recursive = artSpec.Recursive
	options.FollowLink = artSpec.FollowLink
	// Collect
	return artifact.CollectFileArtifact(name, path, options)
}

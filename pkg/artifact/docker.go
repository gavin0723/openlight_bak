// Author: lipixun
// Created Time : 日 12/18 16:35:49 2016
//
// File Name: docker.go
// Description:
//

package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// ArtifactTypeDockerImage defines the dockerimage artifact name
	ArtifactTypeDockerImage = "DockerImage"
)

const (
	dockerImageFileName = "dockerimage.json"
)

// Ensure the interface is implemented
var _ Artifact = (*DockerImageArtifact)(nil)

// DockerImageArtifact implements Aritfact interface for docker image
type DockerImageArtifact struct {
	Path       string `json:"-" yaml:"-"`
	Repository string `json:"repository" yaml:"repository"`
	Image      string `json:"image" yaml:"image"`
	Tag        string `json:"tag" yaml:"tag"`
}

// NewDockerImageArtifact creates a new DockerImageArtifact
func NewDockerImageArtifact(path, repository, image, tag string) *DockerImageArtifact {
	return &DockerImageArtifact{path, repository, image, tag}
}

// NewNewDockerImageArtifactFromPath creates a new DockerImageArtifact from path (which is dumped by dump method)
func NewNewDockerImageArtifactFromPath(p string) (*DockerImageArtifact, error) {
	file, err := os.Open(filepath.Join(p, dockerImageFileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Json decode
	var artifact DockerImageArtifact
	if err := json.NewDecoder(file).Decode(&artifact); err != nil {
		return nil, err
	}
	return &artifact, nil
}

// GetType returns the artifact type
func (artifact *DockerImageArtifact) GetType() string {
	return ArtifactTypeDockerImage
}

// GetPath returns the (original) path of this artifact
func (artifact *DockerImageArtifact) GetPath() string {
	return artifact.Path
}

// String returns the string
func (artifact *DockerImageArtifact) String() string {
	return fmt.Sprintf("%s: %s", artifact.GetType(), artifact.FullName())
}

// FullName returns the image fullname
func (artifact *DockerImageArtifact) FullName() string {
	if artifact.Repository != "" {
		return fmt.Sprintf("%v/%v:%v", artifact.Repository, artifact.Image, artifact.Tag)
	}
	return fmt.Sprintf("%v:%v", artifact.Image, artifact.Tag)
}

// LatestName returns the image latest name
func (artifact *DockerImageArtifact) LatestName() string {
	if artifact.Repository != "" {
		return fmt.Sprintf("%v/%v:latest", artifact.Repository, artifact.Image)
	}
	return fmt.Sprintf("%v:latest", artifact.Image)
}

// Dump will write "dockerimage.json" under its path attribute
func (artifact *DockerImageArtifact) Dump() error {
	file, err := os.Create(filepath.Join(artifact.Path, dockerImageFileName))
	if err != nil {
		return err
	}
	defer file.Close()
	// Json encode
	return json.NewEncoder(file).Encode(artifact)
}

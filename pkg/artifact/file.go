// Author: lipixun
// Created Time : 日 12/18 16:35:39 2016
//
// File Name: file.go
// Description:
//

package artifact

import (
	"fmt"
)

const (
	// ArtifactTypeFile defines the artifact type file
	ArtifactTypeFile = "File"
)

// Ensure the interface is implemented
var _ Artifact = (*FileArtifact)(nil)

// FileArtifact implements Artifact interface for file
type FileArtifact struct {
	Path string `json:"path" yaml:"path"`
}

// NewFileArtifact creates a new FileArtifact
func NewFileArtifact(path string) *FileArtifact {
	return &FileArtifact{path}
}

// GetType returns the artifact type
func (artifact *FileArtifact) GetType() string {
	return ArtifactTypeFile
}

// GetPath returns the (original) path of this artifact
func (artifact *FileArtifact) GetPath() string {
	return artifact.Path
}

// String returns the string
func (artifact *FileArtifact) String() string {
	return fmt.Sprintf("%v: %v", artifact.GetType(), artifact.Path)
}

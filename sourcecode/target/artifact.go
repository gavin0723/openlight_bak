// Author: lipixun
// Created Time : 四 10/20 09:53:50 2016
//
// File Name: artifact.go
// Description:
//
package target

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

const (
	ArtifactTypeFile        = "file"
	ArtifactTypeDockerImage = "dockerImage"

	ArtifactFileAttrFileName     = "filename"
	ArtifactFileAttrRelativePath = "relativePath"

	ArtifactDockerImageAttrFullname   = "fullname"
	ArtifactDockerImageAttrRepository = "repository"
	ArtifactDockerImageAttrImageName  = "imageName"
	ArtifactDockerImageAttrTag        = "tag"
)

type Artifact struct {
	Name  string            `json:"name" yaml:"name"`
	Type  string            `json:"type" yaml:"type"`
	Uri   string            `json:"uri" yaml:"uri"`
	Attrs map[string]string `json:"attrs" yaml:"attrs"`
}

func (this *Artifact) String() string {
	return fmt.Sprintf("%s:%s --> %s", this.Type, this.Name, this.Uri)
}

// Create a standard file artifact
func NewFileArtifact(relativePath string, basePath string) *Artifact {
	return &Artifact{
		Type: ArtifactTypeFile,
		Name: relativePath,
		Uri:  filepath.Join(basePath, relativePath),
		Attrs: map[string]string{
			ArtifactFileAttrFileName:     filepath.Base(relativePath),
			ArtifactFileAttrRelativePath: relativePath,
		},
	}
}

type FileArtifactCollector struct {
	Recursive bool
	Excludes  []*regexp.Regexp
}

func NewFileArtifactCollector(recursive bool, excludes []*regexp.Regexp) *FileArtifactCollector {
	return &FileArtifactCollector{
		Recursive: recursive,
		Excludes:  excludes,
	}
}

// Collect a path
//	Recursive collect will not collect directories
// 	Non-Recursive collect will collect files and directories
func (this *FileArtifactCollector) Collect(p string) ([]*Artifact, error) {
	var artifacts []*Artifact
	if this.Recursive {
		// Recursively collect
		err := filepath.Walk(p, func(subPath string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				// Add this node
				relativePath, err := filepath.Rel(p, subPath)
				if err != nil {
					return err
				}
				// Check exclude
				if !this.isExcluded(relativePath) {
					artifacts = append(artifacts, this.collectPath(relativePath, p))
				}
			}
			// Done
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Not recursively collect
		infos, err := ioutil.ReadDir(p)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		for _, info := range infos {
			relativePath := info.Name()
			if !this.isExcluded(relativePath) {
				artifacts = append(artifacts, this.collectPath(relativePath, p))
			}

		}
	}
	// Done
	return artifacts, nil
}

// Collect thie path
// 	Parameter p is the relative path
func (this *FileArtifactCollector) collectPath(p string, basePath string) *Artifact {
	return NewFileArtifact(p, basePath)
}

// Check if the path should be excluded
func (this *FileArtifactCollector) isExcluded(p string) bool {
	excluded := false
	for _, exp := range this.Excludes {
		if exp.MatchString(p) {
			// Exclude
			excluded = true
			break
		}
	}
	return excluded
}

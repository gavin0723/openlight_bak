// Author: lipixun
// Created Time : 四 10/20 10:26:09 2016
//
// File Name: config.go
// Description:
//	The config used by cli
package spec

import (
	"fmt"
)

const (
	DefaultRepositoryType = "git"
)

type Repository struct {
	Uri      string              // The repository uri
	Source   string              // The source uri
	Local    RepositoryLocalInfo // The local info
	Metadata RepositoryMetadata  // The metadata
	Spec     *RepositorySpec     // The repository spec
}

func (this *Repository) GetTargetKey(name string) string {
	return GetTargetKey(name, this)
}

func (this *Repository) String() string {
	return fmt.Sprintf("%s --> %s @@ %s", this.Uri, this.Metadata.String(), this.Local.String())
}

type RepositoryLocalInfo struct {
	Path string
}

func (this *RepositoryLocalInfo) String() string {
	return this.Path
}

type RepositoryMetadata struct {
	Branch  string `json:"branch"`
	Commit  string `json:"commit"`
	Message string `json:"message"`
}

func (this *RepositoryMetadata) String() string {
	return fmt.Sprintf("%s@%s<--[%s]", this.Commit, this.Branch, this.Message)
}

// The repository spec
type RepositorySpec struct {
	Uri     string `yaml:"uri"`
	Options struct {
		Default struct {
			Build struct {
				Target string `yaml:"target"`
			} `yaml:"build"`
		} `yaml:"default"`
	} `yaml:"options"`
	References map[string]*RepositoryReferenceSpec `yaml:"references"` // Key is repository uri
	Targets    map[string]*TargetSpec              `yaml:"targets"`    // Key is target name
}

type RepositoryReferenceSpec struct {
	Remote string `yaml:"remote"` // The repository remote path, either a local path or url
	Branch string `yaml:"branch"`
	Commit string `yaml:"commit"`
	Finder struct {
		Type   string                 `yaml:"type"`
		Params map[string]interface{} `yaml:"params"`
	} `yaml:"finder"` // The repository finder
}

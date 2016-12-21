// Author: lipixun
// Created Time : 五 10/28 11:20:41 2016
//
// File Name: target.go
// Description:
//	The spec
package spec

import (
	"fmt"
	"path/filepath"
)

type Target struct {
	Name       string
	Repository *Repository
	Spec       *TargetSpec
}

// Get the unique key of target
func (this *Target) Key() string {
	return fmt.Sprintf("%s:%s", this.Repository.Uri, this.Name)
}

// Get the absolute path of this target
func (this *Target) Path() string {
	return filepath.Join(this.Repository.Local.Path, this.Spec.Path)
}

func GetTargetKey(target string, r *Repository) string {
	return fmt.Sprintf("%s:%s", r.Uri, target)
}

type TargetSpec struct {
	Path  string `yaml:"path"` // The relative path of the target in the repository
	Build struct {
		Type   string           `yaml:"type"` // The build type of the target
		Shell  *ShellBuildSpec  `yaml:"shell"`
		Docker *DockerBuildSpec `yaml:"docker"`
		Golang *GolangBuildSpec `yaml:"golang"`
		Python *PythonBuildSpec `yaml:"python"`
	} `yaml:"build"`
	Deps map[string]*TargetDependencySpec `yaml:"deps"` // The key is target dependency name
}

type TargetDependencySpec struct {
	Target     string `yaml:"target"`     // The target name
	Repository string `yaml:"repository"` // The repository uri, empty means the repository itself
	Options    struct {
		Build bool `yaml:"build"` // Build the dependency or not
	} `yaml:"options"` // The dependency options
}

func (this *TargetDependencySpec) Key() string {
	return fmt.Sprintf("%s:%s", this.Repository, this.Target)
}

func GetTargetDependencyKey(name, target, repository string) string {
	return fmt.Sprintf("%s:%s:%s", name, repository, target)
}

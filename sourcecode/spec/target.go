// Author: lipixun
// Created Time : 五 10/28 11:20:41 2016
//
// File Name: target.go
// Description:
//	The spec
package spec

import (
	"github.com/ops-openlight/openlight/uri"
)

type Target struct {
	SourceCode *struct {
		Type   string            `yaml:"type"`
		Golang *GolangSourceCode `yaml:"golang"`
		Python *PythonSourceCode `yaml:"python"`
	} `yaml:"sourcecode"`
	Builder *struct {
		Type     string           `yaml:"type"`
		Makefile *MakefileBuilder `yaml:"makefile"`
		Shell    *ShellBuilder    `yaml:"shell"`
		Docker   *DockerBuilder   `yaml:"docker"`
		Golang   *GolangBuilder   `yaml:"golang"`
		Python   *PythonBuilder   `yaml:"python"`
	} `yaml:"builder"`
	Deps []TargetDependency `yaml:"deps"`
}

type TargetDependency struct {
	Target     string `json:"target" yaml:"target"`
	Repository *struct {
		Uri    string             `json:"uri" yaml:"uri"`       // The uri of repository
		Remote *uri.RepositoryUri `json:"remote" yaml:"remote"` // The remote of this repository
	} `json:"repository" yaml:"repository"`
}

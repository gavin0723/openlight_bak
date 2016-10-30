// Author: lipixun
// Created Time : 二 10/18 22:03:36 2016
//
// File Name: shell.go
// Description:
//	Openlight shell target
//
// 	Build
//		Shell target will be built by shell script file
// 		Before execute the script, builder will inject the following variables into environment:
// 			OPS_Build_OutputPath 		The output directory (abspath)
// 			OPS_Build_Tag 				The build tag
// 			OPS_Build_Time 				The build time, in RFC3339 format
package target

import (
	"errors"
	"github.com/ops-openlight/openlight/sourcecode/spec"
)

const (
	ShellBuilderLogHeader = "ShellBuilder"
	BuilderTypeShell      = "shell"
)

type ShellBuilder struct {
	builderSpec *spec.ShellBuilder
}

func NewShellBuilder(builderSpec *spec.ShellBuilder) (Builder, error) {
	// Create new shell builder
	// Check config
	if builderSpec == nil {
		return nil, errors.New("Require shell builder spec")
	}
	// Create a new shell builder
	return &ShellBuilder{builderSpec: builderSpec}, nil
}

func ShellBuilderCreator(s *spec.Target) (Builder, error) {
	// Create new shell builder
	// Check config
	if s.Builder == nil || s.Builder.Shell == nil {
		return nil, errors.New("Require shell builder spec")
	}
	// Create a new shell builder
	return NewShellBuilder(s.Builder.Shell)
}

func (this *ShellBuilder) Type() string {
	return BuilderTypeShell
}

func (this *ShellBuilder) Build(ctx *TargetBuildContext) (*BuildResult, error) {
	// Build this target
	return nil, errors.New("Not implemented")
}

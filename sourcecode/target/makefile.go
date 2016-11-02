// Author: lipixun
// Created Time : 二 10/18 22:02:50 2016
//
// File Name: makefile.go
// Description:
//	Openlight makefile target
//
// 	Build
//		Shell target will be built by make command
// 		Before execute the make command, builder will inject the following variables by -e:
// 			CI_OUTPUT 			The output directory (abspath)
// 			CI_BRANCH 			The build branch
// 			CI_TAG 				The build tag
// 			CI_TIME 			The build time, in RFC3339 format
// 			CI_COMMIT 			The build commit id
// 			CI_BUILDGRAPH 		The build graph
//
package target

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/sourcecode/spec"
	"github.com/satori/go.uuid"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	MakefileLogHeader   = "MakefileBuilder"
	BuilderTypeMakefile = "makefile"
)

type MakefileBuilder struct {
	builderSpec *spec.MakefileBuilder
}

func NewMakefileBuilder(builderSpec *spec.MakefileBuilder) (Builder, error) {
	// Create a new makefile target
	return &MakefileBuilder{builderSpec: builderSpec}, nil
}

func MakefileBuilderCreator(s *spec.Target) (Builder, error) {
	// Create a new makefile target
	if s.Builder != nil {
		return NewMakefileBuilder(s.Builder.Makefile)
	} else {
		return NewMakefileBuilder(nil)
	}
}

func (this *MakefileBuilder) Type() string {
	return BuilderTypeMakefile
}

func (this *MakefileBuilder) Build(ctx *TargetBuildContext) (*BuildResult, error) {
	// Build makefile target
	if len(ctx.Target.Deps) > 0 {
		// Print a warning
		ctx.Workspace.Logger.WriteWarningHeaderln(MakefileLogHeader, "Makefile builder will not build dependency")
	}
	// Get the output path
	outputPath, err := ctx.Workspace.FileSystem.GetGeneratePath(hex.EncodeToString(uuid.NewV4().Bytes()), true)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Failed to ensure generate path, error: ", err))
	}
	// Create the commands
	var args []string
	var dirName string // The directory which makefile in (abspath)
	// Check makefile filename
	if this.builderSpec != nil && this.builderSpec.File != "" {
		dirName = filepath.Join(ctx.Target.Repository.Local.RootPath, filepath.Dir(this.builderSpec.File))
		makefile := filepath.Base(this.builderSpec.File)
		if makefile != "makefile" && makefile != "Makefile" {
			args = append(args, "-f", makefile)
		}
	} else {
		dirName = ctx.Target.Repository.Local.RootPath
	}
	// Check variables
	if this.builderSpec != nil {
		for name, value := range this.builderSpec.Variables {
			args = append(args, "-e", fmt.Sprintf("%s=%s", name, value))
		}
	}
	// Inject build parameters
	args = append(args,
		"-e", fmt.Sprintf("CI_OUTPUT=%s", outputPath),
		"-e", fmt.Sprintf("CI_BRANCH=%s", ctx.Target.Repository.Metadata.Branch),
		"-e", fmt.Sprintf("CI_TAG=%s", ctx.Option.Tag),
		"-e", fmt.Sprintf("CI_TIME=%s", ctx.Option.Time.Format(time.RFC3339)),
		"-e", fmt.Sprintf("CI_COMMIT=%s", ctx.Target.Repository.Metadata.Commit),
	)
	// Set the target
	if this.builderSpec != nil && this.builderSpec.Target != "" {
		args = append(args, this.builderSpec.Target)
	}
	// Create the command
	cmd := exec.Command("make", args...)
	cmd.Dir = dirName
	if ctx.Workspace.Verbose() {
		// Connect stdout and stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Ignore the stderr and stdout
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	// Run and wait
	if ctx.Workspace.Verbose() {
		ctx.Workspace.Logger.WriteInfoHeaderln(MakefileLogHeader, "Run exec command: make ", strings.Join(args, " "))
	}
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	// Collect file artifact
	var excludes []*regexp.Regexp
	if this.builderSpec != nil {
		for _, exclude := range this.builderSpec.Output.Collect.Excludes {
			exp, err := regexp.Compile(exclude)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to compile exclude pattern [%s], error: %s", exclude, err))
			}
			excludes = append(excludes, exp)
		}
	}
	recursive := false
	if this.builderSpec != nil {
		recursive = this.builderSpec.Output.Collect.Recursive
	}
	collector := NewFileArtifactCollector(recursive, excludes)
	artifacts, err := collector.Collect(outputPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to collect artifacts, error: %s", err))
	}
	// Done
	return &BuildResult{
		OutputPath: outputPath,
		Metadata:   BuildMetadata{Tag: ctx.Option.Tag, Time: ctx.Option.Time},
		Artifacts:  artifacts,
	}, nil
}

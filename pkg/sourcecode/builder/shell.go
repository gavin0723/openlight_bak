// Author: lipixun
// Created Time : 二 10/18 22:02:50 2016
//
// File Name: makefile.go
// Description:
//	Openlight makefile target
//
// 	Build
//		Shell target will be built by make command
//
package builder

import (
	"errors"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	ShellLogHeader = "Shell"

	BuilderTypeShell = "shell"
)

type ShellSourceCodeBuilder struct{}

func NewShellSourceCodeBuilder() *ShellSourceCodeBuilder {
	return new(ShellSourceCodeBuilder)
}

// Create new environment for the builder
func (this *ShellSourceCodeBuilder) NewEnviron(builder *Builder) (Environment, error) {
	return NewGeneralEnvironment(filepath.Join(builder.EnvironmentPath(), BuilderTypeShell))
}

// Prepare for the target
func (this *ShellSourceCodeBuilder) Prepare(target *spec.Target, env Environment, context *BuilderContext) error {
	environ := env.(*GeneralEnvironment)
	if environ == nil {
		return errors.New("Invalid environment")
	}
	shellSpec := target.Spec.Build.Shell
	if shellSpec == nil {
		return errors.New("Shell build spec not defined")
	}
	path, err := environ.EnsureTargetPath(target)
	if err != nil {
		return err
	}
	// Link
	for _, link := range shellSpec.Links {
		if err := GeneralLink(target, &link, path); err != nil {
			return err
		}
	}
	// Done
	return nil
}

func (this *ShellSourceCodeBuilder) Build(target *spec.Target, env Environment, context *BuilderContext) error {
	startBuildTime := time.Now()
	shellSpec := target.Spec.Build.Shell
	if shellSpec == nil {
		return errors.New("Shell build spec not defined")
	}
	if len(shellSpec.Collectors) == 0 {
		return errors.New("No artifact collector defined in shell build spec")
	}
	logger := context.Workspace.Logger.GetLoggerWithHeader(ShellLogHeader)
	// The output path
	outputPath, err := context.Builder.EnsureTargetOutputPath(target)
	if err != nil {
		return err
	}
	// Create shell build command
	var workDir string
	if shellSpec.WorkDir == "" {
		workDir = target.Path()
	}
	// Get the environment variables
	environVars := GetBuildMetadataEnvironVars(
		outputPath,
		target.Repository.Metadata.Branch,
		target.Repository.Metadata.Commit,
		context.Builder.Options.Tag,
		context.Builder.Options.Time,
	)
	// Create the command
	var args []string
	args = append(args, shellSpec.Args...)
	cmd := exec.Command(shellSpec.Command, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), environVars...)
	if context.Workspace.Verbose {
		// Connect stdout and stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Ignore the stderr and stdout
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	// Run shell command
	logger.LeveledPrintf(log.LevelDebug, "Run command: %s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		return err
	}
	// Collect the artifacts
	artifacts, err := CollectFileArtifactBySpecs(outputPath, shellSpec.Collectors)
	if err != nil {
		return err
	}
	// Create the build result
	buildResult := spec.NewBuildResult(target, context.Builder.NewBuildMetadata(target))
	buildResult.Metadata.Builder = BuilderTypeShell
	buildResult.Metadata.BuildTimeUsage = time.Now().Sub(startBuildTime).Seconds()
	buildResult.Metadata.LinkedPath = env.GetTargetPath(target)
	buildResult.Metadata.OutputPath = outputPath
	for _, art := range artifacts {
		buildResult.Artifacts[art.GetName()] = art
	}
	context.Builder.SetBuildResultDependency(target, buildResult)
	context.Builder.AddResult(target, buildResult)
	// Done
	return nil
}

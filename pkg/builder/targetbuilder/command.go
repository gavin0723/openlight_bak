// Author: lipixun
// Created Time : 二 10/18 22:02:50 2016
//
// File Name: makefile.go
// Description:
//
//	Set environment variables:
//
//		* CI_OUTPUT
//		* CI_BRANCH
//		* CI_COMMIT
//		* CI_BUILD_TIME
//		* CI_COMMIT_TIME
//		* CI_TAG
//

package targetbuilder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/builder/buildcontext"
	"github.com/ops-openlight/openlight/pkg/repository"
	"github.com/ops-openlight/openlight/pkg/utils"
)

// CommandTargetBuilder builds command target
type CommandTargetBuilder struct {
	target *repository.Target
	spec   *pbSpec.CommandTarget
}

// NewCommandTargetBuilder creates a new CommandTargetBuilder
func NewCommandTargetBuilder(target *repository.Target, spec *pbSpec.CommandTarget) (*CommandTargetBuilder, error) {
	if target == nil {
		return nil, errors.New("Require target")
	}
	if spec == nil {
		return nil, errors.New("Require spec")
	}
	return &CommandTargetBuilder{target, spec}, nil
}

// Build the target
func (builder *CommandTargetBuilder) Build(ctx buildcontext.Context) (artifact.Artifact, error) {
	if builder.spec.GetCommand() == "" {
		return nil, errors.New("Require command")
	}
	// The output path
	outputPath, err := ctx.GetTargetOutputDir(builder.target, true)
	if err != nil {
		log.Errorf("Failed to get target output dir: %v", err)
		return nil, err
	}
	// The work dir
	workdir := builder.target.Path()
	if builder.spec.Workdir != "" {
		workdir = filepath.Join(workdir, builder.spec.Workdir)
	}
	// Get the environment variables
	gitRepoInfo, err := utils.GetGitRepositoryInfo(builder.target.Path())
	if err != nil {
		log.Warnf("Failed to get git repository info: %v", err)
	}
	envs := append(os.Environ(),
		fmt.Sprintf("CI_OUTPUT=%v", outputPath),
		fmt.Sprintf("CI_BUILD_TIME=%v", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("CI_TAG=%v", ctx.Tag()),
	)
	if gitRepoInfo != nil {
		envs = append(envs,
			fmt.Sprintf("CI_BRANCH=%v", gitRepoInfo.Branch),
			fmt.Sprintf("CI_COMMIT=%v", gitRepoInfo.Commit),
			fmt.Sprintf("CI_COMMIT_TIME=%v", gitRepoInfo.CommitTime),
		)
	}
	for _, env := range builder.spec.GetEnvs() {
		envs = append(envs, env)
	}
	// Create the command
	cmd := exec.Command(builder.spec.Command, builder.explainArgs(envs, builder.spec.Args)...)
	cmd.Dir = workdir
	cmd.Env = envs
	// Show debug
	if ctx.Verbose() {
		log.Debugln("CommandTargetBuilder.Build: Environment variables:\n\t", strings.Join(envs, "\n\t"))
		log.Debugln("CommandTargetBuilder.Build: Run command:", strings.Join(cmd.Args, " "))
	}
	// Run command
	if ctx.Verbose() {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("Failed to run command: %v", err)
		}
	} else {
		if outBytes, err := cmd.CombinedOutput(); err != nil {
			log.Errorln("Failed to run command:\n", string(outBytes))
			return nil, fmt.Errorf("Failed to run command: %v", err)
		}
	}

	// Create the artifact
	return artifact.NewFileArtifact(outputPath), nil
}

func (builder *CommandTargetBuilder) explainArgs(envs []string, args []string) []string {
	envMap := make(map[string]string)
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 1 {
			envMap[parts[0]] = ""
		} else {
			envMap[parts[0]] = parts[1]
		}
	}
	// Explain
	var explainedArgs []string
	for _, arg := range args {
		explainedArgs = append(explainedArgs, os.Expand(arg, func(key string) string { return envMap[key] }))
	}
	// Done
	return explainedArgs
}

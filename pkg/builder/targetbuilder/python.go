// Author: lipixun
// Created Time : 五 10/28 11:35:07 2016
//
// File Name: python.go
// Description:
//
//	Set environment variables:
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

const (
	_DefaultPythonSetupScriptFile = "setup.py"
)

// PythonLibTargetBuilder builds python libraray target through python setup file
type PythonLibTargetBuilder struct {
	target *repository.Target
	spec   *pbSpec.PythonLibTarget
}

// NewPythonLibTargetBuilder creates a new PythonLibTargetBuilder
func NewPythonLibTargetBuilder(target *repository.Target, spec *pbSpec.PythonLibTarget) (*PythonLibTargetBuilder, error) {
	if target == nil {
		return nil, errors.New("Require target")
	}
	if spec == nil {
		return nil, errors.New("Require spec")
	}
	return &PythonLibTargetBuilder{target, spec}, nil
}

// Build the target
func (builder *PythonLibTargetBuilder) Build(ctx buildcontext.Context) (artifact.Artifact, error) {
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
	log.Debugln("PythonLibTargetBuilder.Build: Workdir:", workdir)
	// Get the environment variables
	gitRepoInfo, err := utils.GetGitRepositoryInfo(builder.target.Path())
	if err != nil {
		log.Warnf("Failed to get git repository info: %v", err)
	}
	envs := []string{
		fmt.Sprintf("CI_OUTPUT=\"%v\"", outputPath),
		fmt.Sprintf("CI_BUILD_TIME=\"%v\"", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("CI_TAG=\"%v\"", ctx.Tag()),
	}
	if gitRepoInfo != nil {
		envs = append(envs,
			fmt.Sprintf("CI_BRANCH=\"%v\"", gitRepoInfo.Branch),
			fmt.Sprintf("CI_COMMIT=\"%v\"", gitRepoInfo.Commit),
			fmt.Sprintf("CI_COMMIT_TIME=\"%v\"", gitRepoInfo.CommitTime),
		)
	}
	// Get the script file
	scriptFile := builder.spec.GetSetup()
	if scriptFile == "" {
		scriptFile = _DefaultPythonSetupScriptFile
	}

	// Create command
	cmd := exec.Command("python", scriptFile, "sdist", "-d", outputPath)
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), envs...)

	// Run command
	log.Debugln("PythonLibTargetBuilder.Build: Run command:", strings.Join(cmd.Args, " "))
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

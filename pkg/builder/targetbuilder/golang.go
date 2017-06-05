// Author: lipixun
// Created Time : 四 10/20 21:50:40 2016
//
// File Name: golang.go
// Description:
//	Golang builder
// 		Will inject the following variables:
// 			- buildBranch 		The build branch
// 			- buildCommit 		The build commit
//			- buildCommitTime 	The build commit time
// 			- buildTime 		The build time in RFC3339 format
// 			- buildTag 			The build tag
//

package targetbuilder

import (
	"errors"
	"fmt"
	"io"
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

// GoBinaryTargetBuilder builds go binary target
type GoBinaryTargetBuilder struct {
	target *repository.Target
	spec   *pbSpec.GoBinaryTarget
}

// NewGoBinaryTargetBuilder creates a new GoBinaryTargetBuilder
func NewGoBinaryTargetBuilder(target *repository.Target, spec *pbSpec.GoBinaryTarget) (*GoBinaryTargetBuilder, error) {
	if target == nil {
		return nil, errors.New("Require target")
	}
	if spec == nil {
		return nil, errors.New("Require spec")
	}
	return &GoBinaryTargetBuilder{target, spec}, nil
}

// Build the target
func (builder *GoBinaryTargetBuilder) Build(ctx buildcontext.Context) (artifact.Artifact, error) {
	if builder.spec.GetPackage() == "" {
		return nil, errors.New("Require package")
	}
	// The output path
	outputPath, err := ctx.GetTargetOutputDir(builder.target, true)
	if err != nil {
		log.Errorf("Failed to get target output dir: %v", err)
		return nil, err
	}

	// Create go build command
	gitRepoInfo, err := utils.GetGitRepositoryInfo(builder.target.Path())
	if err != nil {
		log.Warnf("Failed to get git repository info: %v", err)
	}
	var ldflags string
	if gitRepoInfo != nil {
		ldflags = fmt.Sprintf(
			"-X main.buildBranch=%v -X main.buildCommit=%v -X main.buildCommitTime=%v -X main.buildTime=%v -X main.buildTag=%v",
			gitRepoInfo.Branch,
			gitRepoInfo.Commit,
			gitRepoInfo.CommitTime,
			time.Now().Format(time.RFC3339),
			ctx.Tag(),
		)
	} else {
		ldflags = fmt.Sprintf(
			"-X main.buildTime=%v -X main.buildTag=%v",
			time.Now().Format(time.RFC3339),
			ctx.Tag(),
		)
	}
	outputName := builder.spec.Output
	if outputName == "" {
		outputName = builder.spec.Package[strings.LastIndex(builder.spec.Package, "/")+1:]
	}

	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", filepath.Join(outputPath, outputName), builder.spec.Package)
	cmd.Dir = builder.target.Path()
	if len(builder.spec.Envs) > 0 {
		cmd.Env = append(os.Environ(), builder.spec.Envs...)
	}

	// Run command
	log.Debugln("GoBinaryTargetBuilder.Build: Run command =", cmd.Path, strings.Join(cmd.Args, " "))
	if ctx.Verbose() {
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("Failed to run command: %v", err)
		}
	} else {
		if outBytes, err := cmd.CombinedOutput(); err != nil {
			log.Errorln("Failed to run command:\n", string(outBytes))
			return nil, fmt.Errorf("Failed to run command: %v", err)
		}
	}

	// Copy to install path
	if builder.spec.Install {
		log.Debugln("GoBinaryTargetBuilder.Build: Install binary")
		goPath, err := builder.getGoPath()
		if err != nil {
			log.Errorln("Failed to get GOPATH:", err)
			return nil, fmt.Errorf("Failed to get GOPATH: %v", err)
		}
		if goPath == "" {
			log.Errorln("GOPATH not found, cannot install the binary")
			return nil, errors.New("GOPATH not found, cannot install the binary")
		}
		// Copy the binary
		rfile, err := os.Open(filepath.Join(outputPath, outputName))
		if err != nil {
			log.Errorln("Failed to open output binary file:", err)
			return nil, fmt.Errorf("Failed to open output binary file: %v", err)
		}
		defer rfile.Close()
		wfile, err := os.OpenFile(filepath.Join(goPath, "bin", outputName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			log.Errorln("Failed to create installing binary file:", err)
			return nil, fmt.Errorf("Failed to create installing binary file: %v", err)
		}
		defer wfile.Close()
		// Copy it
		_, err = io.Copy(wfile, rfile)
		if err != nil {
			log.Errorln("Failed to copy file:", err)
			return nil, fmt.Errorf("Failed to copy file: %v", err)
		}
	}

	// Done
	return artifact.NewFileArtifact(outputPath), nil
}

func (builder *GoBinaryTargetBuilder) getGoPath() (string, error) {
	for _, s := range os.Environ() {
		if strings.HasPrefix(s, "GOPATH=") {
			return utils.GetRealPath(s[7:])
		}
	}
	return "", nil
}

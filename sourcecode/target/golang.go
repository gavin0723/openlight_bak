// Author: lipixun
// Created Time : 四 10/20 21:50:40 2016
//
// File Name: golang.go
// Description:
//	Golang builder
// 		Will inject the following variables:
// 			- buildBranch 	The build branch
// 			- buildCommit 	The build commit
// 			- buildTime 	The build time in RFC3339 format
// 			- buildTag 		The build tag
//			- buildGraph 	The build graph json string
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
	"strings"
	"time"
)

const (
	GolangLogHeader = "Golang"

	BuilderTypeGolang = "golang"

	SourceCodeTypeGolang = "golang"
)

type GolangBuilder struct {
	builderSpec *spec.GolangBuilder
}

func NewGolangBuilder(builderSpec *spec.GolangBuilder) (Builder, error) {
	// Create a new golang builder
	if builderSpec == nil {
		return nil, errors.New("Require golang builder spec")
	}
	// Done
	return &GolangBuilder{builderSpec: builderSpec}, nil
}

func GolangBuilderCreator(s *spec.Target) (Builder, error) {
	// Create a new golang builder
	if s.Builder == nil || s.Builder.Golang == nil {
		return nil, errors.New("Require golang builder spec")
	}
	// Done
	return NewGolangBuilder(s.Builder.Golang)
}

func (this *GolangBuilder) Type() string {
	return BuilderTypeGolang
}

func (this *GolangBuilder) Build(ctx *TargetBuildContext) (*BuildResult, error) {
	// Build golang
	if this.builderSpec.Package == "" {
		return nil, errors.New("Require package to build")
	}
	// We need link go packages from target and the targets which this target is depend on
	// So, we have to initialize a go path and then link all packages
	outputPath, err := ctx.Workspace.FileSystem.GetGeneratePath(hex.EncodeToString(uuid.NewV4().Bytes()), true)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Failed to ensure generate path, error: ", err))
	}
	// Create the go path
	goPath := filepath.Join(outputPath, "golang")
	goSrcPath := filepath.Join(goPath, "src")
	if err := os.MkdirAll(goSrcPath, os.ModePerm); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create path [%s], error: %s", goSrcPath, err))
	}
	// Link packages
	if err := this.linkTarget(ctx.Target, goSrcPath); err != nil {
		return nil, errors.New(fmt.Sprintf("Link packages error: %s", err))
	}
	// Create go build command
	var args []string = []string{"build"}
	// The output path
	if this.builderSpec.Output != "" {
		args = append(args, "-o", filepath.Join(outputPath, this.builderSpec.Output))
	} else {
		args = append(args, "-o", outputPath)
	}
	// The build metadata
	args = append(args, "-ldflags", fmt.Sprintf("-X main.buildBranch=%s -X main.buildCommit=%s -X main.buildTime=%s -X main.buildTag=%s",
		ctx.Target.Repository.Metadata.Branch,
		ctx.Target.Repository.Metadata.Commit,
		ctx.Option.Time.Format(time.RFC3339),
		ctx.Option.Tag,
	))
	// Add the build package
	args = append(args, this.builderSpec.Package)
	// Create the command
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToLower(e), "gopath=") {
			env = append(env, e)
		}
	}
	env = append(env, fmt.Sprintf("GOPATH=%s", goPath))
	cmd := exec.Command("go", args...)
	cmd.Env = env
	if ctx.Workspace.Verbose() {
		// Connect stdout and stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Ignore the stderr and stdout
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	// Run go build
	if ctx.Workspace.Verbose() {
		ctx.Workspace.Logger.WriteInfoHeaderln(GolangLogHeader, "Run exec command: go ", strings.Join(args, " "))
	}
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	// Good, create the artifact
	var binaryFilename, binaryFileRelativePath string
	if this.builderSpec.Output != "" {
		binaryFilename = filepath.Base(this.builderSpec.Output)
		binaryFileRelativePath = this.builderSpec.Output
	} else {
		binaryFilename = filepath.Base(this.builderSpec.Package)
		binaryFileRelativePath = binaryFilename
	}
	if _, err := os.Stat(filepath.Join(outputPath, binaryFileRelativePath)); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to check output binary file [%s], error: %s", filepath.Join(outputPath, binaryFileRelativePath), err))
	}
	// Done
	return &BuildResult{
		OutputPath: outputPath,
		Metadata:   BuildMetadata{Tag: ctx.Option.Tag, Time: ctx.Option.Time},
		Artifacts: []*Artifact{
			NewFileArtifact(binaryFileRelativePath, outputPath),
		},
	}, nil
}

// Link target
func (this *GolangBuilder) linkTarget(target *Target, goSourcePath string) error {
	// Check the target source code
	if target.SourceCode != nil && target.SourceCode.Type() == SourceCodeTypeGolang {
		// Golang source code, link it
		golangSourceCode, ok := target.SourceCode.(GolangSourceCode)
		if !ok {
			return errors.New(fmt.Sprintf("Cannot cast sourcecode to golang sourcecode for target [%s]", target.Key()))
		}
		for _, pkg := range golangSourceCode.GetPackages(target) {
			pkgSourcePath := filepath.Join(goSourcePath, pkg.Package)
			if _, err := os.Stat(pkgSourcePath); err == nil {
				// Found the package source path
				return errors.New(fmt.Sprintf("Package source path [%s] already exists for target [%s]", pkgSourcePath, target.Key()))
			} else if !os.IsNotExist(err) {
				// Error
				return errors.New(fmt.Sprintf("Check package source path [%s] for target [%s] error: %s", pkgSourcePath, target.Key(), err))
			} else {
				// Not found, link the path
				if err := os.MkdirAll(filepath.Dir(pkgSourcePath), os.ModePerm); err != nil {
					return errors.New(fmt.Sprintf("Failed to prepare path [%s], error: %s", filepath.Dir(pkgSourcePath), err))
				}
				if err := os.Symlink(pkg.Path, pkgSourcePath); err != nil {
					return errors.New(fmt.Sprintf("Failed to link package source path [%s] from [%s] for target [%s] error: %s", pkgSourcePath, pkg.Path, target.Key(), err))
				}
			}
		}
	}
	// Link all deps
	for _, dep := range target.Deps {
		if err := this.linkTarget(dep.Target, goSourcePath); err != nil {
			return err
		}
	}
	// Done
	return nil
}

// The golang sourcecode common interface
type GolangSourceCode interface {
	GetPackages(target *Target) []GolangPackage
}

// The standard golang source code implementation
type StdGolangSourceCode struct {
	srcSpec *spec.GolangSourceCode
}

func NewGolangSourceCode(srcSpec *spec.GolangSourceCode) (SourceCode, error) {
	// Create a new golang sourcecode
	if srcSpec == nil {
		return nil, errors.New("Require golang sourcecode spec")
	}
	// Done
	return &StdGolangSourceCode{srcSpec: srcSpec}, nil
}

func GolangSourceCodeCreator(s *spec.Target) (SourceCode, error) {
	// Create a new golang sourcecode
	if s.SourceCode == nil || s.SourceCode.Golang == nil {
		return nil, errors.New("Require golang sourcecode spec")
	}
	// Done
	return NewGolangSourceCode(s.SourceCode.Golang)
}

func (this *StdGolangSourceCode) Type() string {
	return SourceCodeTypeGolang
}

type GolangPackage struct {
	Package string // The package name
	Path    string // The package path (absoluate path)
}

// Get packages
func (this *StdGolangSourceCode) GetPackages(target *Target) []GolangPackage {
	var pkgs []GolangPackage
	for _, pkg := range this.srcSpec.Packages {
		pkgs = append(pkgs, GolangPackage{Package: pkg.Package, Path: filepath.Join(target.Repository.Local.RootPath, pkg.Path)})
	}
	return pkgs
}

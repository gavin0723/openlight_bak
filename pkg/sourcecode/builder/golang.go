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
package builder

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	GolangLogHeader = "Golang"

	BuilderTypeGolang = "golang"
)

type GolangSourceCodeBuilder struct{}

func NewGolangSourceCodeBuilder() *GolangSourceCodeBuilder {
	return new(GolangSourceCodeBuilder)
}

// Create new environment for the builder
func (this *GolangSourceCodeBuilder) NewEnviron(builder *Builder) (Environment, error) {
	return NewGolangEnvironment(filepath.Join(builder.EnvironmentPath(), BuilderTypeGolang))
}

// Prepare for the target
func (this *GolangSourceCodeBuilder) Prepare(target *spec.Target, env Environment, context *BuilderContext) error {
	environ := env.(*GolangEnvironment)
	if environ == nil {
		return errors.New("Invalid environment")
	}
	packagePath, err := environ.EnsurePackagePath(target)
	if err != nil {
		return err
	}
	golangSpec := target.Spec.Build.Golang
	if golangSpec == nil {
		return errors.New("Golang build spec not defined")
	}
	// Link
	if !golangSpec.NoVendor {
		if err := this.tryLinkVendor(target.Path(), packagePath); err != nil {
			return err
		}
	}
	for _, link := range golangSpec.Links {
		if err := GeneralLink(target, &link, packagePath); err != nil {
			return err
		}
	}
	// Done
	return nil
}

func (this *GolangSourceCodeBuilder) tryLinkVendor(sourcePath, targetPath string) error {
	names := []string{"vendor", "Godeps"}
	for _, name := range names {
		path := filepath.Join(sourcePath, name)
		if _, err := os.Stat(path); err == nil {
			if err := os.Symlink(path, filepath.Join(targetPath, name)); err != nil {
				return errors.New(fmt.Sprintf("Failed to link [%s], error: %s", name, err))
			}
		}
	}
	// Done
	return nil
}

// Build the target
func (this *GolangSourceCodeBuilder) Build(target *spec.Target, env Environment, context *BuilderContext) error {
	startBuildTime := time.Now()
	golangSpec := target.Spec.Build.Golang
	if golangSpec == nil {
		return errors.New("Golang build spec not defined")
	}
	logger := context.Workspace.Logger.GetLoggerWithHeader(GolangLogHeader)
	// Create go build command
	var args []string = []string{"build"}
	// The output path
	outputPath, err := context.Builder.EnsureTargetOutputPath(target)
	if err != nil {
		return err
	}
	// The build metadata
	args = append(args, "-ldflags", fmt.Sprintf("-X main.buildBranch=%s -X main.buildCommit=%s -X main.buildTime=%s -X main.buildTag=%s",
		target.Repository.Metadata.Branch,
		target.Repository.Metadata.Commit,
		context.Builder.Options.Time.Format(time.RFC3339),
		context.Builder.Options.Tag,
	))
	// Add the build package
	buildPackage := golangSpec.BuildPackage
	if buildPackage == "" {
		buildPackage = golangSpec.Package
	}
	// The output
	if golangSpec.Output != "" {
		args = append(args, "-o", filepath.Join(outputPath, golangSpec.Output))
	} else {
		names := strings.Split(buildPackage, "/")
		args = append(args, "-o", filepath.Join(outputPath, names[len(names)-1]))
	}
	// The build package
	args = append(args, buildPackage)
	// Create the command
	var environVars []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToLower(e), "gopath=") {
			environVars = append(environVars, e)
		}
	}
	environVars = append(environVars, fmt.Sprintf("GOPATH=%s", env.Path()))
	cmd := exec.Command("go", args...)
	cmd.Dir = env.Path()
	cmd.Env = environVars
	if context.Workspace.Verbose {
		// Connect stdout and stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Ignore the stderr and stdout
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	// Run go build
	logger.LeveledPrintf(log.LevelDebug, "Run command: %s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		return err
	}
	// Good, create the artifact
	artifactName := golangSpec.Name
	if artifactName == "" {
		artifactName = BuilderDefaultArtifactName
	}
	// Collect the artifact
	artifact, err := artifact.CollectFileArtifact(artifactName, outputPath, artifact.NewDefaultCollectFileArtifactOptions())
	if err != nil {
		return err
	}
	// Create the build result
	buildResult := spec.NewBuildResult(target, context.Builder.NewBuildMetadata(target))
	buildResult.Metadata.Builder = BuilderTypeGolang
	buildResult.Metadata.BuildTimeUsage = time.Now().Sub(startBuildTime).Seconds()
	buildResult.Metadata.LinkedPath = env.GetTargetPath(target)
	buildResult.Metadata.OutputPath = outputPath
	buildResult.Artifacts[artifact.GetName()] = artifact
	context.Builder.SetBuildResultDependency(target, buildResult)
	context.Builder.AddResult(target, buildResult)
	// Done
	return nil
}

type GolangEnvironment struct {
	path    string
	targets map[string]*GolangTargetEnvironment
}

type GolangTargetEnvironment struct {
	Target  *spec.Target
	Package string
	Path    string
}

func NewGolangEnvironment(path string) (Environment, error) {
	if err := os.MkdirAll(filepath.Join(path, "src"), os.ModePerm); err != nil {
		return nil, err
	}
	return &GolangEnvironment{
		path:    path,
		targets: make(map[string]*GolangTargetEnvironment),
	}, nil
}

func (this *GolangEnvironment) Path() string {
	return this.path
}

func (this *GolangEnvironment) GetTargets() []*spec.Target {
	var targets []*spec.Target
	for _, environ := range this.targets {
		targets = append(targets, environ.Target)
	}
	return targets
}

func (this *GolangEnvironment) GetTargetPath(target *spec.Target) string {
	environ := this.targets[target.Key()]
	if environ != nil {
		return environ.Path
	}
	return ""
}

func (this *GolangEnvironment) GetEnvironVars() map[string]string {
	vars := make(map[string]string)
	for _, environ := range this.targets {
		vars[GetBuildTargetEnvironVarKey(environ.Target)] = environ.Path
	}
	return vars
}

func (this *GolangEnvironment) EnsurePackagePath(target *spec.Target) (string, error) {
	environ := this.targets[target.Key()]
	if environ != nil {
		return environ.Path, nil
	}
	golangSpec := target.Spec.Build.Golang
	if golangSpec == nil {
		return "", errors.New("Golang build spec not defined")
	}
	packageName := golangSpec.Package
	if packageName == "" {
		return "", errors.New("Golang build package name not defined")
	}
	// Make dirs
	packagePath := filepath.Join(this.path, "src", packageName)
	if err := os.MkdirAll(packagePath, os.ModePerm); err != nil {
		return "", err
	}
	// Create the environment and return
	environ = &GolangTargetEnvironment{Target: target, Package: packageName, Path: packagePath}
	this.targets[target.Key()] = environ
	return packagePath, nil
}

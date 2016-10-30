// Author: lipixun
// Created Time : 五 10/28 11:35:07 2016
//
// File Name: python.go
// Description:
//	Python target, this target is used to build python binary via nuitka
//  The normal python pip package could be simply built by makefile target or shell target
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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	PythonLogHeader = "Python"

	BuilderTypePython = "python"

	SourceCodeTypePython = "python"

	PythonCodeInsertPointLine = "# [BUILD INSERT POINT]"
)

type PythonBuilder struct {
	builderSpec *spec.PythonBuilder
}

func NewPythonBuilder(builderSpec *spec.PythonBuilder) (Builder, error) {
	// Create a new python builder
	if builderSpec == nil {
		return nil, errors.New("Require python builder spec")
	}
	// Done
	return &PythonBuilder{builderSpec: builderSpec}, nil
}

func PythonBuilderCreator(s *spec.Target) (Builder, error) {
	// Create a new python builder
	if s.Builder == nil || s.Builder.Python == nil {
		return nil, errors.New("Require python builder spec")
	}
	// Done
	return NewPythonBuilder(s.Builder.Python)
}

func (this *PythonBuilder) Type() string {
	return BuilderTypePython
}

func (this *PythonBuilder) Build(ctx *TargetBuildContext) (*BuildResult, error) {
	// Build python
	if this.builderSpec.EntryScriptFile == "" {
		return nil, errors.New("Require entry script file")
	}
	// We need link python packages from target and the targets which this target is depend on
	// So, we have to initialize a python path and then link all packages
	outputPath, err := ctx.Workspace.FileSystem.GetGeneratePath(hex.EncodeToString(uuid.NewV4().Bytes()), true)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Failed to ensure generate path, error: ", err))
	}
	pythonPath := filepath.Join(outputPath, "python")
	// Create the python path
	if err := os.MkdirAll(pythonPath, os.ModePerm); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create path [%s], error: %s", pythonPath, err))
	}
	// Link packages
	if err := this.linkTarget(ctx.Target, pythonPath); err != nil {
		return nil, errors.New(fmt.Sprintf("Link packages error: %s", err))
	}
	// List the modules
	infos, err := ioutil.ReadDir(pythonPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read python path [%s], error: %s", pythonPath, err))
	}
	var modules []string
	for _, info := range infos {
		modules = append(modules, info.Name())
	}
	// Create the entry script
	scriptbytes, err := ioutil.ReadFile(filepath.Join(ctx.Target.Repository.Local.RootPath, this.builderSpec.EntryScriptFile))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load entry script file [%s], error: %s", filepath.Join(ctx.Target.Repository.Local.RootPath, this.builderSpec.EntryScriptFile), err))
	}
	// Find the insert point
	script := string(scriptbytes)
	if idx := strings.Index(script, PythonCodeInsertPointLine); idx != -1 {
		// Insert the build metadata
		script = strings.Replace(
			script,
			PythonCodeInsertPointLine,
			fmt.Sprintf(
				"\n# Openlight build tool generated metadata\nbuildBranch=\"%s\"\nbuildCommit=\"%s\"\nbuildTime=\"%s\"\nbuildTag=\"%s\"\n# -*- ------------------------------ -*-\n# \n",
				ctx.Target.Repository.Metadata.Branch,
				ctx.Target.Repository.Metadata.Commit,
				ctx.Option.Time.Format(time.RFC3339),
				ctx.Option.Tag,
			),
			1,
		)
	}
	outScriptFile := "__noname__.py"
	outBinaryFile := "__noname__.exe"
	if err := ioutil.WriteFile(filepath.Join(outputPath, outScriptFile), []byte(script), os.ModePerm); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to generate entry script to [%s], error: %s", filepath.Join(outputPath, outScriptFile), err))
	}
	// Create nuitka build command
	var args []string = []string{"--remove-output"}
	for _, module := range modules {
		args = append(args, "--recurse-to", module)
	}
	args = append(args, outScriptFile)
	// Create the command
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToLower(e), "pythonpath=") {
			env = append(env, e)
		}
	}
	env = append(env, fmt.Sprintf("PYTHONPATH=%s", pythonPath))
	cmd := exec.Command("nuitka", args...)
	cmd.Env = env
	cmd.Dir = outputPath
	if ctx.Workspace.Verbose() {
		// Connect stdout and stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Ignore the stderr and stdout
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	// Run nuitka
	if ctx.Workspace.Verbose() {
		ctx.Workspace.Logger.WriteInfoHeaderln(PythonLogHeader, "Run exec command: nuitka ", strings.Join(args, " "))
	}
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var outBinaryName string
	if this.builderSpec.Output != "" {
		outBinaryName = this.builderSpec.Output
	} else {
		outBinaryName = ctx.Target.Name
	}
	// Good, create the artifact
	if err := os.Rename(filepath.Join(outputPath, outBinaryFile), filepath.Join(outputPath, outBinaryName)); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to rename [%s] to [%s], error: %s", filepath.Join(outputPath, outBinaryFile), filepath.Join(outputPath, outBinaryName), err))
	}
	if _, err := os.Stat(filepath.Join(outputPath, outBinaryName)); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to check output binary file [%s], error: %s", filepath.Join(outputPath, outBinaryName), err))
	}
	// Done
	return &BuildResult{
		OutputPath: outputPath,
		Metadata:   BuildMetadata{Tag: ctx.Option.Tag, Time: ctx.Option.Time},
		Artifacts: []*Artifact{
			NewFileArtifact(outBinaryName, outputPath),
		},
	}, nil
}

// Link target
func (this *PythonBuilder) linkTarget(target *Target, pythonPath string) error {
	// Check the target source code
	if target.SourceCode != nil && target.SourceCode.Type() == SourceCodeTypePython {
		// Python source code, link it
		pySourceCode, ok := target.SourceCode.(PythonSourceCode)
		if !ok {
			return errors.New(fmt.Sprintf("Cannot cast sourcecode to golang sourcecode for target [%s]", target.Key()))
		}
		for _, module := range pySourceCode.GetModules(target) {
			modulePath := filepath.Join(pythonPath, module.Module)
			if _, err := os.Stat(modulePath); err == nil {
				// Found the module source path
				return errors.New(fmt.Sprintf("Module path [%s] already exists for target [%s]", modulePath, target.Key()))
			} else if !os.IsNotExist(err) {
				// Error
				return errors.New(fmt.Sprintf("Check module path [%s] for target [%s] error: %s", modulePath, target.Key(), err))
			} else {
				// Not found, link the path
				if err := os.Symlink(module.Path, modulePath); err != nil {
					return errors.New(fmt.Sprintf("Failed to link module path [%s] from [%s] for target [%s] error: %s", modulePath, module.Path, target.Key(), err))
				}
			}
		}
	}
	// Link all deps
	for _, dep := range target.Deps {
		if err := this.linkTarget(dep.Target, pythonPath); err != nil {
			return err
		}
	}
	// Done
	return nil
}

// The golang sourcecode common interface
type PythonSourceCode interface {
	GetModules(target *Target) []PythonModule
}

// The standard golang source code implementation
type StdPythonSourceCode struct {
	srcSpec *spec.PythonSourceCode
}

func NewPythonSourceCode(srcSpec *spec.PythonSourceCode) (SourceCode, error) {
	// Create a new golang sourcecode
	if srcSpec == nil {
		return nil, errors.New("Require golang sourcecode spec")
	}
	// Done
	return &StdPythonSourceCode{srcSpec: srcSpec}, nil
}

func PythonSourceCodeCreator(s *spec.Target) (SourceCode, error) {
	// Create a new golang sourcecode
	if s.SourceCode == nil || s.SourceCode.Python == nil {
		return nil, errors.New("Require golang sourcecode spec")
	}
	// Done
	return NewPythonSourceCode(s.SourceCode.Python)
}

func (this *StdPythonSourceCode) Type() string {
	return SourceCodeTypePython
}

type PythonModule struct {
	Module string // The module name
	Path   string // The module path (absoluate path)
}

// Get packages
func (this *StdPythonSourceCode) GetModules(target *Target) []PythonModule {
	var modules []PythonModule
	for _, module := range this.srcSpec.Modules {
		modules = append(modules, PythonModule{Module: module.Module, Path: filepath.Join(target.Repository.Local.RootPath, module.Path)})
	}
	return modules
}

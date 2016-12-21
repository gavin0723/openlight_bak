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
//
package builder

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/artifact"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	PythonLogHeader            = "Python"
	PythonNuitkaLogHeader      = "Python.Nuitka"
	PythonSetupScriptLogHeader = "Python.SetupScript"

	BuilderTypePython = "python"

	PythonBuildTypeScript = "script"
	PythonBuildTypeNuitka = "nuitka"

	PythonCodeInsertPointLine = "# [BUILD INSERT POINT]"

	DefaultPythonSetupScriptFile    = "setup.py"
	DefaultPythonSetupScriptCommand = "sdist"

	PythonNuitkaBuildTypeBinary = "binary"
	PythonNuitkaBuildTypeLib    = "lib"
)

type PythonSourceCodeBuilder struct{}

func NewPythonSourceCodeBuilder() *PythonSourceCodeBuilder {
	return new(PythonSourceCodeBuilder)
}

// Create new environment for the builder
func (this *PythonSourceCodeBuilder) NewEnviron(builder *Builder) (Environment, error) {
	return NewPythonEnvironment(filepath.Join(builder.EnvironmentPath(), BuilderTypePython))
}

// Prepare for the target
func (this *PythonSourceCodeBuilder) Prepare(target *spec.Target, env Environment, context *BuilderContext) error {
	environ := env.(*PythonEnvironment)
	if environ == nil {
		return errors.New("Invalid environment")
	}
	pythonSpec := target.Spec.Build.Python
	if pythonSpec == nil {
		return errors.New("Python build spec not defined")
	}
	path, err := environ.EnsurePackagePath(target)
	if err != nil {
		return err
	}
	// Do general link
	for _, link := range pythonSpec.Links {
		if err := GeneralLink(target, &link, path); err != nil {
			return err
		}
	}
	// Done
	return nil
}

// Build the target
func (this *PythonSourceCodeBuilder) Build(target *spec.Target, env Environment, context *BuilderContext) error {
	pythonSpec := target.Spec.Build.Python
	if pythonSpec == nil {
		return errors.New("Python build spec not defined")
	}
	// Build
	if pythonSpec.Type == PythonBuildTypeScript {
		return this.runScriptBuild(target, env, context)
	} else if pythonSpec.Type == PythonBuildTypeNuitka {
		return this.runNuitkaBuild(target, env, context)
	} else {
		return errors.New(fmt.Sprintf("Unknown python build type [%s]", pythonSpec.Type))
	}
}

// Run python setup build
func (this *PythonSourceCodeBuilder) runScriptBuild(target *spec.Target, env Environment, context *BuilderContext) error {
	startBuildTime := time.Now()
	pythonSpec := target.Spec.Build.Python
	if pythonSpec == nil {
		return errors.New("Python build spec not defined")
	}
	scriptSpec := pythonSpec.Script
	// Get the environment
	environ := env.(*PythonEnvironment)
	if environ == nil {
		return errors.New("Invalid environment")
	}
	logger := context.Workspace.Logger.GetLoggerWithHeader(PythonSetupScriptLogHeader)
	// Get the script file
	scriptFile := DefaultPythonSetupScriptFile
	if scriptSpec != nil && scriptSpec.ScriptFile != "" {
		scriptFile = scriptSpec.ScriptFile
	}
	// Get the command
	command := DefaultPythonSetupScriptCommand
	if scriptSpec != nil && scriptSpec.Command != "" {
		command = scriptSpec.Command
	}
	// The source path
	sourcePath := env.GetTargetPath(target)
	if sourcePath == "" {
		return errors.New("Source path not found")
	}
	// The output path
	outputPath, err := context.Builder.EnsureTargetOutputPath(target)
	if err != nil {
		return err
	}
	// Create python script build command
	var args []string = []string{scriptFile, command, "-d", outputPath}
	// Add the environment variables
	var environVars []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToLower(e), "pythonpath=") {
			environVars = append(environVars, e)
		}
	}
	// Add PYTHONPATH
	environVars = append(environVars, fmt.Sprintf("PYTHONPATH=%s", environ.GetPythonPathVar()))
	// Add build metadata
	environVars = append(environVars, GetBuildMetadataEnvironVars(
		outputPath,
		target.Repository.Metadata.Branch,
		target.Repository.Metadata.Commit,
		context.Builder.Options.Tag,
		context.Builder.Options.Time,
	)...)
	// Create the command
	cmd := exec.Command("python", args...)
	cmd.Dir = sourcePath
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
	// Collect the artifacts in the output directory
	artifactName := pythonSpec.Name
	if artifactName == "" {
		artifactName = BuilderDefaultArtifactName
	}
	art, err := artifact.CollectFileArtifact(artifactName, outputPath, artifact.NewDefaultCollectFileArtifactOptions())
	if err != nil {
		return err
	}
	// Create the build result
	buildResult := spec.NewBuildResult(target, context.Builder.NewBuildMetadata(target))
	buildResult.Metadata.Builder = BuilderTypePython
	buildResult.Metadata.BuildTimeUsage = time.Now().Sub(startBuildTime).Seconds()
	buildResult.Metadata.LinkedPath = sourcePath
	buildResult.Metadata.OutputPath = outputPath
	buildResult.Artifacts[art.GetName()] = art
	context.Builder.SetBuildResultDependency(target, buildResult)
	context.Builder.AddResult(target, buildResult)
	// Done
	return nil
}

// Run python nuitka build
func (this *PythonSourceCodeBuilder) runNuitkaBuild(target *spec.Target, env Environment, context *BuilderContext) error {
	startBuildTime := time.Now()
	pythonSpec := target.Spec.Build.Python
	if pythonSpec == nil {
		return errors.New("Python build spec not defined")
	}
	if pythonSpec.Nuitka == nil {
		return errors.New("Python nuitka build spec not defined")
	}
	nuitkaSpec := pythonSpec.Nuitka
	if nuitkaSpec.Type != PythonNuitkaBuildTypeBinary && nuitkaSpec.Type != PythonNuitkaBuildTypeLib {
		return errors.New(fmt.Sprintf("Unknown python nuitka build type [%s]", nuitkaSpec.Type))
	}
	// Get the environment
	environ := env.(*PythonEnvironment)
	if environ == nil {
		return errors.New("Invalid environment")
	}
	logger := context.Workspace.Logger.GetLoggerWithHeader(PythonNuitkaLogHeader)
	// The source path
	sourcePath := env.GetTargetPath(target)
	if sourcePath == "" {
		return errors.New("Source path not found")
	}
	// The output path
	outputPath, err := context.Builder.EnsureTargetOutputPath(target)
	if err != nil {
		return err
	}
	if nuitkaSpec.Output == "" {
		return errors.New("Require nuitka output")
	}
	// Prepare the command args
	var args []string = []string{"--output-dir", outputPath}
	for _, module := range nuitkaSpec.Modules {
		args = append(args, "--recurse-to", module)
	}
	// Prepare the environment variables
	var environVars []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToLower(e), "pythonpath=") {
			environVars = append(environVars, e)
		}
	}
	// Add PYTHONPATH
	environVars = append(environVars, fmt.Sprintf("PYTHONPATH=%s", environ.GetPythonPathVar()))
	// Add build metadata
	environVars = append(environVars, GetBuildMetadataEnvironVars(
		outputPath,
		target.Repository.Metadata.Branch,
		target.Repository.Metadata.Commit,
		context.Builder.Options.Tag,
		context.Builder.Options.Time,
	)...)
	// Check it's a binary build or a lib build
	var nuitkaOutputFile, buildOutputFile string
	if nuitkaSpec.Type == PythonNuitkaBuildTypeBinary {
		// Build as binary
		if nuitkaSpec.Binary == nil || nuitkaSpec.Binary.EntryScript == "" {
			return errors.New("Entry script not defined")
		}
		if nuitkaSpec.Output == "" {
			return errors.New("Output not defined")
		}
		buildOutputFile = nuitkaSpec.Output
		entryScript := nuitkaSpec.Binary.EntryScript
		// Load the entry script and try to write the metadata
		entryScriptFile := filepath.Join(sourcePath, entryScript)
		entryScriptData, err := ioutil.ReadFile(entryScriptFile)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to load entry script file [%s], error: %s", entryScriptFile, err))
		}
		// Find the insert point
		script := string(entryScriptData)
		if idx := strings.Index(script, PythonCodeInsertPointLine); idx != -1 {
			// Find the insert point, generate a new entry script with the metadata
			script = strings.Replace(
				script,
				PythonCodeInsertPointLine,
				fmt.Sprintf(
					"\n# Openlight build tool generated metadata\nbuildBranch=\"%s\"\nbuildCommit=\"%s\"\nbuildTime=\"%s\"\nbuildTag=\"%s\"\n# -*- ------------------------------ -*-\n# \n",
					target.Repository.Metadata.Branch,
					target.Repository.Metadata.Commit,
					context.Builder.Options.Time.Format(time.RFC3339),
					context.Builder.Options.Tag,
				),
				1,
			)
			// Write the new entry script
			newEntryScript := "__noname_entry__.py"
			newEntryScriptFile := filepath.Join(sourcePath, newEntryScript)
			nuitkaOutputFile = "__noname_entry__.exe"
			if err := ioutil.WriteFile(newEntryScriptFile, []byte(script), os.ModePerm); err != nil {
				return errors.New(fmt.Sprintf("Failed to generate entry script to [%s], error: %s", newEntryScriptFile, err))
			}
			entryScript = newEntryScript
		} else {
			// Use the original entry script
			nuitkaOutputFile = fmt.Sprintf("%s.exe", entryScript)
		}
		// Add entry script to args
		args = append(args, entryScript)
	} else {
		// Binary as library
		args = append(args, "--module")
		// Not implemented
		return errors.New("Build as a python module via nuitka is not supported yet")
	}
	// Run the command
	cmd := exec.Command("nuitka", args...)
	cmd.Env = environVars
	cmd.Dir = sourcePath
	// Run nuitka
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
	if context.Workspace.Verbose {
		logger.LeveledPrintf(log.LevelDebug, "Run command: %s %s\n", strings.Join(cmd.Args, " "))
		logger.LeveledPrintf(log.LevelDebug, "Environment Variables: %s\n", strings.Join(environVars, ";"))
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	// Rename the output file
	nuitkaOutputFile = filepath.Join(outputPath, nuitkaOutputFile)
	buildOutputFile = filepath.Join(outputPath, buildOutputFile)
	if err := os.Rename(nuitkaOutputFile, buildOutputFile); err != nil {
		return err
	}
	// Collect the artifacts in the output directory
	artifactName := pythonSpec.Name
	if artifactName == "" {
		artifactName = BuilderDefaultArtifactName
	}
	art, err := artifact.CollectFileArtifact(artifactName, outputPath, artifact.NewDefaultCollectFileArtifactOptions())
	if err != nil {
		return err
	}
	// Create the build result
	buildResult := spec.NewBuildResult(target, context.Builder.NewBuildMetadata(target))
	buildResult.Metadata.Builder = BuilderTypePython
	buildResult.Metadata.BuildTimeUsage = time.Now().Sub(startBuildTime).Seconds()
	buildResult.Metadata.LinkedPath = sourcePath
	buildResult.Metadata.OutputPath = outputPath
	buildResult.Artifacts[art.GetName()] = art
	context.Builder.SetBuildResultDependency(target, buildResult)
	context.Builder.AddResult(target, buildResult)
	// Done
	return nil
}

type PythonEnvironment struct {
	path    string
	targets map[string]*PythonTargetEnvironment
}

type PythonTargetEnvironment struct {
	Target      *spec.Target
	Path        string
	ModulePaths []string
}

func NewPythonEnvironment(path string) (Environment, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}
	return &PythonEnvironment{
		path:    path,
		targets: make(map[string]*PythonTargetEnvironment),
	}, nil
}

func (this *PythonEnvironment) Path() string {
	return this.path
}

func (this *PythonEnvironment) GetTargets() []*spec.Target {
	var targets []*spec.Target
	for _, environ := range this.targets {
		targets = append(targets, environ.Target)
	}
	return targets
}

func (this *PythonEnvironment) GetTargetPath(target *spec.Target) string {
	environ := this.targets[target.Key()]
	if environ != nil {
		return environ.Path
	}
	return ""
}

func (this *PythonEnvironment) GetEnvironVars() map[string]string {
	vars := make(map[string]string)
	for _, environ := range this.targets {
		vars[GetBuildTargetEnvironVarKey(environ.Target)] = environ.Path
	}
	return vars
}

func (this *PythonEnvironment) GetPythonPathVar() string {
	var paths []string
	for _, environ := range this.targets {
		for _, _path := range environ.ModulePaths {
			paths = append(paths, _path)
		}
	}
	return strings.Join(paths, ":")
}

func (this *PythonEnvironment) EnsurePackagePath(target *spec.Target) (string, error) {
	environ := this.targets[target.Key()]
	if environ == nil {
		pythonSpec := target.Spec.Build.Python
		if pythonSpec == nil {
			return "", errors.New("Python build spec not defined")
		}
		path := filepath.Join(this.path, GetTargetRegularKey(target))
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return "", err
		}
		modulePaths := pythonSpec.ModulePaths
		if len(modulePaths) == 0 {
			modulePaths = []string{target.Path()}
		}
		environ = &PythonTargetEnvironment{
			Target:      target,
			Path:        path,
			ModulePaths: modulePaths,
		}
		this.targets[target.Key()] = environ
	}
	// Done
	return environ.Path, nil
}

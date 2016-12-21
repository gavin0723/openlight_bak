// Author: lipixun
// Created Time : 五 10/28 11:35:25 2016
//
// File Name: python.go
// Description:
//	Python spec
package spec

type PythonBuildSpec struct {
	Name        string                 `yaml:"name"`        // The name of this build, the generated artifact will be use the same name
	Type        string                 `yaml:"type"`        // The build type, either script or nuitka
	Links       []SourceCodeLink       `yaml:"links"`       // The target to link into the package
	ModulePaths []string               `yaml:"modulePaths"` // The path to the parent of the module, not the module itself. This path will be added to PYTHONPATH. Will use the target path if not specified
	Script      *PythonSetupBuildSpec  `yaml:"script"`      // The python script build spec
	Nuitka      *PythonNuitkaBuildSpec `yaml:"nuitka"`      // The python nuitka build spec
}

// The python setup script build spec
type PythonSetupBuildSpec struct {
	ScriptFile string `yaml:"script"`  // The setup script file, "setup.py" by default
	Command    string `yaml:"command"` // The command of the script, "sdist" by default
}

// The spec used to build via nuitka
type PythonNuitkaBuildSpec struct {
	Type    string                       `yaml:"type"`    // The build type, either binary or lib
	Binary  *PythonNuitkaBinaryBuildSpec `yaml:"binary"`  // The binary build spec
	Lib     *PythonNuitkaLibBuildSpac    `yaml:"lib"`     // The lib build spec
	Modules []string                     `yaml:"modules"` // The module names to build within nuitka
	Output  string                       `yaml:"output"`  // The output binary name
}

type PythonNuitkaBinaryBuildSpec struct {
	EntryScript string `yaml:"entry"` // The script file as entry file, this is a must
}

type PythonNuitkaLibBuildSpac struct {
}

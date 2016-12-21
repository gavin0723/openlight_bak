// Author: lipixun
// Created Time : 日 10/23 20:40:53 2016
//
// File Name: shell.go
// Description:
//
package spec

type ShellBuildSpec struct {
	Command    string                                `yaml:"command"`    // The shell script command
	WorkDir    string                                `yaml:"workDir"`    // The work directory path of the script. Will use the directory of the target
	Links      []SourceCodeLink                      `yaml:"links"`      // The target to link into the package
	Args       []string                              `yaml:"args"`       // The arguments to run the script file
	Collectors map[string]*FileArtifactCollectorSpec `yaml:"collectors"` // The file artifact collector spec
}

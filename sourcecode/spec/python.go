// Author: lipixun
// Created Time : 五 10/28 11:35:25 2016
//
// File Name: python.go
// Description:
//	Python spec
package spec

type PythonBuilder struct {
	EntryScriptFile string `yaml:"entryScriptFile" json:"entryScriptFile"` // The entry script filename
	Output          string `yaml:"output" json:"output"`                   // The output binary name
}

type PythonSourceCode struct {
	Modules []struct {
		Module string `yaml:"module" json:"module"` // The module name
		Path   string `yaml:"path" json:"path"`     // The path
	} `yaml:"modules" json:"modules"`
}

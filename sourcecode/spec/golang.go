// Author: lipixun
// Created Time : 日 10/23 20:40:33 2016
//
// File Name: golang.go
// Description:
//
package spec

type GolangBuilder struct {
	Package string `yaml:"package"`
	Output  string `yaml:"output"`
}

type GolangSourceCode struct {
	Packages []struct {
		Package string `yaml:"package"`
		Path    string `yaml:"path"`
	} `yaml:"packages"`
}

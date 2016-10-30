// Author: lipixun
// Created Time : 五 10/28 11:35:25 2016
//
// File Name: python.go
// Description:
//	Python spec
package spec

type PythonBuilder struct {
}

type PythonSourceCode struct {
	Packages struct {
		Package string `yaml:"package"`
		Path    string `yaml:"path"`
	} `yaml:"packages"`
}

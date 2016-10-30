// Author: lipixun
// Created Time : 日 10/23 20:40:46 2016
//
// File Name: makefile.go
// Description:
//
package spec

type MakefileBuilder struct {
	Target    string            `yaml:"target"`    // The make target
	File      string            `yaml:"file"`      // The makefile filename
	Variables map[string]string `yaml:"variables"` // The additional variables (which will be set by -e)
}

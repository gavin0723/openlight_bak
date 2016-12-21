// Author: lipixun
// Created Time : 日 10/23 20:40:33 2016
//
// File Name: golang.go
// Description:
//
package spec

type GolangBuildSpec struct {
	Name         string           `yaml:"name"`         // The name of the artifact (build result)
	Package      string           `yaml:"package"`      // The top package name
	Links        []SourceCodeLink `yaml:"links"`        // The target to link into the package
	BuildPackage string           `yaml:"buildPackage"` // The package to build, if not specified will use package field
	NoVendor     bool             `yaml:"noVendor"`     // Do not link vendor package
	Output       string           `yaml:"output"`       // The build output file name, will use the last part of the build package if not specifed
}

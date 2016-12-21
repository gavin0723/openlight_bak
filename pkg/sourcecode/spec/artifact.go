// Author: lipixun
// Created Time : 一 12/19 13:38:16 2016
//
// File Name: artifact.go
// Description:
//	The artifact
package spec

type FileArtifactCollectorSpec struct {
	Path       string `yaml:"path"`
	Recursive  bool   `yaml:"recursive"`
	FollowLink bool   `yaml:"followLink"`
	Includes   string `yaml:"includes"`
	Excludes   string `yaml:"excludes"`
}

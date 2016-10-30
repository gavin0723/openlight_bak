// Author: lipixun
// Created Time : 日 10/23 20:40:27 2016
//
// File Name: docker.go
// Description:
//
package spec

type DockerBuilder struct {
	Repository string     `yaml:"repository"` // The repository name
	Image      string     `yaml:"image"`      // The image name
	TagPrefix  string     `yaml:"tagPrefix"`  // The tag prefix
	Dockerfile string     `yaml:"dockerfile"` // The dockerfile path, Dockerfile by default
	MarkLatest bool       `yaml:"markLatest"` // Mark the built image as "latest"
	Push       bool       `yaml:"push"`       // Push the built image
	NoPull     bool       `yaml:"nopull"`     // Donot pull on build (if necessary images are ready)
	NoCache    bool       `yaml:"nocache"`    // Donot use cache to build this image
	Files      []struct { // The files to add into docker build tar
		Dep *struct {
			Target     string `yaml:"target"`     // The dependent target
			Repository string `yaml:"repository"` // The dependent repository
			Artifact   string `yaml:"artifact"`   // The dependent artifact
		} `yaml:"dep"` // The dependency
		LocalName  string `yaml:"localName"`  // The local filename
		TargetName string `yaml:"targetName"` // The target filename to add into docker build tar
	} `yaml:"files"`
}

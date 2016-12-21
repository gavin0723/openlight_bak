// Author: lipixun
// Created Time : 日 10/23 20:40:27 2016
//
// File Name: docker.go
// Description:
//
package spec

type DockerBuildSpec struct {
	Name       string                `yaml:"name"`       // The artifact name
	Repository string                `yaml:"repository"` // The repository name
	Image      string                `yaml:"image"`      // The image name
	TagPrefix  string                `yaml:"tagPrefix"`  // The tag prefix
	Dockerfile string                `yaml:"dockerfile"` // The dockerfile path, Dockerfile by default
	MarkLatest bool                  `yaml:"markLatest"` // Mark the built image as "latest"
	NoPull     bool                  `yaml:"nopull"`     // Donot pull on build (if necessary images are ready)
	NoCache    bool                  `yaml:"nocache"`    // Donot use cache to build this image
	Files      []DockerBuildFileSpec `yaml:"files"`      // The files/dirs to add into docker build tar
}

type DockerBuildFileSpec struct {
	Target string `yaml:"target"` // The target file / dir name
	Source struct {
		Dep *struct {
			Name     string `yaml:"name"`     // The dependency name
			Artifact string `yaml:"artifact"` // The artifact name
		} `yaml:"dep"` // Get file from dependency
		Local *struct {
			Path string `yaml:"path"` // The local filename
		} `yaml:"local"` // Get file from local
	}
}

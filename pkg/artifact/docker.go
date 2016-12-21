// Author: lipixun
// Created Time : 日 12/18 16:35:49 2016
//
// File Name: docker.go
// Description:
//	The docker artifact
package artifact

import (
	"fmt"
)

const (
	ArtifactTypeDocker = "docker"

	DockerArtifactAttrFullname   = "fullname"
	DockerArtifactAttrRepository = "repository"
	DockerArtifactAttrImage      = "image"
	DockerArtifactAttrTag        = "tag"
)

type DockerArtifact struct {
	Name       string `json:"name" yaml:"name"`         // The name of this artifact
	Fullname   string `json:"fullname" yaml:"fullname"` // The fullname of the image
	Repository string `json:"repository" yaml:"repository"`
	Image      string `json:"image" yaml:"image"`
	Tag        string `json:"tag" yaml:"tag"`
}

func NewDockerArtifact(name, fullname, repository, image, tag string) *DockerArtifact {
	return &DockerArtifact{
		Name:       name,
		Fullname:   fullname,
		Repository: repository,
		Image:      image,
		Tag:        tag,
	}
}

func (this *DockerArtifact) GetName() string {
	return this.Name
}

func (this *DockerArtifact) GetType() string {
	return ArtifactTypeFile
}

func (this *DockerArtifact) GetAttr(name string) interface{} {
	switch name {
	case DockerArtifactAttrFullname:
		return this.Fullname
	case DockerArtifactAttrRepository:
		return this.Repository
	case DockerArtifactAttrImage:
		return this.Image
	case DockerArtifactAttrTag:
		return this.Tag
	default:
		return nil
	}
}

func (this *DockerArtifact) String() string {
	return fmt.Sprintf("%s: %s", ArtifactTypeDocker, this.Fullname)
}

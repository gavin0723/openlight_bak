// Author: lipixun
// Created Time : 五 10/28 11:26:06 2016
//
// File Name: repository.go
// Description:
//	The repository
package repository

import (
	"fmt"
	"github.com/ops-openlight/openlight/sourcecode/spec"
	"github.com/ops-openlight/openlight/uri"
)

const (
	RepositorySpecFileName = ".op.yaml"
)

type Repository struct {
	Reference *uri.RepositoryReference // The reference of this repository
	Metadata  RepositoryMetadata       // The metadata
	Local     RepositoryLocalInfo      // The local info
	Spec      *spec.Repository         // The repository spec
}

type RepositoryMetadata struct {
	Branch  string
	Commit  string
	Message string
}

func (this *RepositoryMetadata) String() string {
	return fmt.Sprintf("%s@%s<--[%s]", this.Commit, this.Branch, this.Message)
}

type RepositoryLocalInfo struct {
	RootPath string
}

func (this *Repository) Uri() string {
	return this.Spec.Uri
}

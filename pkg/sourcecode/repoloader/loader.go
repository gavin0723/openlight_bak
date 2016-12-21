// Author: lipixun
// Created Time : 六 12/10 23:47:17 2016
//
// File Name: loader.go
// Description:
//	The repository loader
package repoloader

import (
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"github.com/ops-openlight/openlight/pkg/workspace"
)

var (
	loaders map[string]Loader = map[string]Loader{
		RepositoryTypeGit: NewGitLoader(),
	}
)

type Loader interface {
	Load(remote string, options LoadOptions, ws *workspace.Workspace) (*spec.Repository, error)
}

type LoadOptions struct {
	Branch string
	Commit string
}

func GetLoader(t string) Loader {
	return loaders[t]
}

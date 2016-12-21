// Author: lipixun
// Created Time : 六 12/10 23:47:35 2016
//
// File Name: git.go
// Description:
//	The git repository loader
package repoloader

import (
	"errors"
	"fmt"
	git "github.com/libgit2/git2go"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"path/filepath"
	"strings"
)

const (
	RepositoryTypeGit = "git"
)

type GitLoader struct {
}

func NewGitLoader() Loader {
	return GitLoader{}
}

func (this GitLoader) Load(remote string, options LoadOptions, ws *workspace.Workspace) (*spec.Repository, error) {
	if strings.HasPrefix(remote, "http://") || strings.HasPrefix(remote, "https://") {
		return nil, errors.New("Not implemented")
	} else {
		// Load from local
		if options.Branch != "" {
			ws.Logger.LeveledPrintf(log.LevelWarn, "Branch will be ignored when load from local path for repository [%s]\n", remote)
		}
		if options.Commit != "" {
			ws.Logger.LeveledPrintf(log.LevelWarn, "Commit will be ignored when load from local path for repository [%s]\n", remote)
		}
		return this.loadFromLocal(remote)
	}
}

// Create repository from a local path (either a local repository or a cloned remote repository)
func (this GitLoader) loadFromLocal(p string) (*spec.Repository, error) {
	// Open git repository
	gitRepo, err := git.OpenRepositoryExtended(p, 0, "")
	if err != nil {
		return nil, err
	}
	defer gitRepo.Free()
	// Load metadata
	var metadata spec.RepositoryMetadata
	headReference, err := gitRepo.Head()
	if err != nil {
		return nil, err
	}
	metadata.Commit = headReference.Target().String()
	headBranch := headReference.Branch()
	if err != nil {
		return nil, err
	}
	metadata.Branch, err = headBranch.Name()
	if err != nil {
		return nil, err
	}
	commit, err := gitRepo.LookupCommit(headReference.Target())
	if err != nil {
		return nil, err
	}
	metadata.Message = strings.Trim(commit.Message(), "\n\r")
	// Load spec
	repoSpec, err := LoadRepositorySpecFromFile(filepath.Join(p, spec.SpecFileName))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load repository spec file, error: %s", err))
	}
	// Verify the spec
	if repoSpec.Uri == "" {
		return nil, errors.New("Invalid repository spec, uri is required")
	}
	// Create the repository
	repo := &spec.Repository{
		Uri:      repoSpec.Uri,
		Source:   p,
		Metadata: metadata,
		Spec:     repoSpec,
		Local: spec.RepositoryLocalInfo{
			Path: filepath.Dir(filepath.Dir(gitRepo.Path())),
		},
	}
	// Done
	return repo, nil
}

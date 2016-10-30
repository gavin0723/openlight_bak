// Author: lipixun
// Created Time : 五 10/28 11:26:23 2016
//
// File Name: manager.go
// Description:
//	The repository manager
package repository

import (
	"errors"
	"fmt"
	git "github.com/libgit2/git2go"
	"github.com/ops-openlight/openlight/helper/iohelper"
	"github.com/ops-openlight/openlight/sourcecode/spec"
	"github.com/ops-openlight/openlight/uri"
	"path/filepath"
	"strings"
)

const (
	RepositoryDefaultBranch = "master"
)

type RepositoryManager struct {
	Local         bool                                // Enable load repository from local
	OnlyLocal     bool                                // Only load repository from local
	UriOverwrites map[string]*uri.RepositoryReference // The repository uri overwrites, key is the repository uri
	Repository    map[string]*Repository              // The loaded repository
}

func NewRepositoryManager(local, onlyLocal bool) *RepositoryManager {
	// New repository manager
	return &RepositoryManager{
		Local:         local,
		OnlyLocal:     onlyLocal,
		UriOverwrites: make(map[string]*uri.RepositoryReference),
		Repository:    make(map[string]*Repository),
	}
}

func (this *RepositoryManager) Load(u string, reference *uri.RepositoryReference, pathFunc func(reference *uri.RepositoryReference) (string, error)) (*Repository, error) {
	// Load repository
	if u != "" {
		if r, ok := this.Repository[u]; ok {
			if reference != nil {
				if err := this.isRepositoryReferenceMatch(r.Reference, reference); err != nil {
					return nil, err
				}
			}
			// Good
			return r, nil
		}
		// Check the overwrites
		if ref, ok := this.UriOverwrites[u]; ok {
			reference = ref
		}
	}
	// Check the reference
	if reference == nil {
		return nil, errors.New("Require reference (Or rewrited by nil reference)")
	}
	if reference.Local == "" && reference.Remote == nil {
		return nil, errors.New(fmt.Sprintf("Invalid reference to uri [%s], require either path or uri", u))
	}
	if reference.Local != "" && reference.Remote != nil {
		return nil, errors.New(fmt.Sprintf("Invalid reference to uri [%s], cannot specify both path and uri", u))
	}
	if reference.Remote != nil {
		if err := reference.Remote.Validate(); err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid reference to uri [%s], failed to validate uri, error: %s", u, err))
		}
	}
	if reference.Local != "" && !this.Local {
		return nil, errors.New("Load repository from local is disabled")
	}
	if reference.Local == "" && this.OnlyLocal {
		return nil, errors.New("Only load repository from local is enabled")
	}
	// Load the repository
	var repo *Repository
	if reference.Local != "" {
		// Load from local
		var err error
		repo, err = this.loadRepositoryFromLocal(reference.Local)
		if err != nil {
			return nil, err
		}
		repo.Reference = reference
	} else {
		// Load from remote
		if pathFunc == nil {
			return nil, errors.New("Require path function when load from remote")
		}
		_, err := pathFunc(reference)
		if err != nil {
			return nil, err
		}
		// TODO: Implement load from remote
		return nil, errors.New("Not implemented error")
	}
	// Verify the uri
	if u != "" && repo.Uri() != u {
		return nil, errors.New(fmt.Sprintf("Mismatch repository uri, expect [%s] actual [%s]", u, repo.Uri()))
	}
	if r, ok := this.Repository[repo.Uri()]; ok {
		if err := this.isRepositoryReferenceMatch(r.Reference, reference); err != nil {
			// Not match
			return nil, err
		} else {
			// Match, use the previous load one
			return r, nil
		}
	}
	// Add to repo
	this.Repository[u] = repo
	// Done
	return repo, nil
}

func (this *RepositoryManager) isRepositoryReferenceMatch(r1, r2 *uri.RepositoryReference) error {
	// Check if two repository references are match
	if r1.Local != "" && r2.Local != "" {
		return errors.New(fmt.Sprintf("Mismatch repository reference, inconsistent path [%s] and [%s]", r1.Local, r2.Local))
	}
	if r1.Remote == nil && r2.Remote != nil || r1.Remote != nil && r2.Remote == nil {
		return errors.New(fmt.Sprintf("Mismatch repository reference, inconsistent uri [has:%s] and [has:%s]", r1.Remote == nil, r2.Remote == nil))
	} else if r1.Remote != nil && r2.Remote != nil {
		if r1.Remote.Uri != r2.Remote.Uri {
			return errors.New(fmt.Sprintf("Mismatch repository reference, inconsistent uri [%s] and [%s]", r1.Remote.Uri, r2.Remote.Uri))
		}
		b1 := r1.Remote.Branch
		if b1 == "" {
			b1 = RepositoryDefaultBranch
		}
		b2 := r2.Remote.Branch
		if b2 == "" {
			b2 = RepositoryDefaultBranch
		}
		if b1 != b2 {
			return errors.New(fmt.Sprintf("Mismatch repository reference, inconsistent branch [%s] and [%s]", b1, b2))
		}
		if r1.Remote.Commit != r2.Remote.Commit {
			return errors.New(fmt.Sprintf("Mismatch repository reference, inconsistent commit [%s] and [%s]", r1.Remote.Commit, r2.Remote.Commit))
		}
	}
	// Good
	return nil
}

func (this *RepositoryManager) loadRepositoryFromLocal(p string) (*Repository, error) {
	// Load repository from local
	// Cannot use relative path
	if !filepath.IsAbs(p) {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot get absoluate path by [%s] error: %s", p, err))
		}
		p = absPath
	}
	// Check the path
	repoPath, err := iohelper.GetRealPath(p)
	if err != nil {
		return nil, err
	}
	// Create the repository
	return this.createRepository(repoPath)
}

func (this *RepositoryManager) createRepository(p string) (*Repository, error) {
	// Create repository from a local path (either a local repository or a cloned remote repository)
	// Open git repository
	gitRepo, err := git.OpenRepositoryExtended(p, 0, "")
	if err != nil {
		return nil, err
	}
	defer gitRepo.Free()
	// Load metadata
	var metadata RepositoryMetadata
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
	repoSpec, err := spec.LoadRepositorySpecFromFile(filepath.Join(p, RepositorySpecFileName))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load repository spec file, error: %s", err))
	}
	// Verify the spec
	if repoSpec.Uri == "" {
		return nil, errors.New("Invalid repository spec, uri is required")
	}
	// Create the repository
	repo := &Repository{
		Metadata: metadata,
		Spec:     repoSpec,
		Local: RepositoryLocalInfo{
			RootPath: filepath.Dir(filepath.Dir(gitRepo.Path())),
		},
	}
	// Done
	return repo, nil
}

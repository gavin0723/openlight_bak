// Author: lipixun
// Created Time : 六 12/10 13:03:21 2016
//
// File Name: git.go
// Description:
//
package dirdetector

import (
	git "github.com/libgit2/git2go"
	"path/filepath"
)

type GitDirDetector struct{}

func newGitDirDetector() GitDirDetector {
	return GitDirDetector{}
}

func (this GitDirDetector) Detect(p string) (string, error) {
	gitRepo, err := git.OpenRepositoryExtended(p, 0, "")
	if err != nil {
		return "", err
	}
	defer gitRepo.Free()
	return filepath.Dir(filepath.Dir(gitRepo.Path())), nil
}

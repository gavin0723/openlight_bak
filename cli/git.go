// Author: lipixun
// Created Time : 四 10/20 21:34:29 2016
//
// File Name: git.go
// Description:
//	Git util
package cli

import (
	"github.com/libgit2/git2go"
	"os"
	"path/filepath"
)

func GetGitRootPath(p string) (string, error) {
	r, err := git.OpenRepositoryExtended(p, 0, "")
	if err != nil {
		return "", err
	}
	defer r.Free()
	// Return the directory of path (the end of the path is .git)
	return filepath.Dir(filepath.Dir(r.Path())), nil
}

func GetGitRootFromCurrentDirectory() (string, error) {
	rootRepoPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	rootRepoPath, err = GetGitRootPath(rootRepoPath)
	if err != nil {
		return "", err
	}
	return rootRepoPath, nil
}

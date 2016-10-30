// Author: lipixun
// Created Time : 五 10/21 16:07:11 2016
//
// File Name: util.go
// Description:
//	Utility
package build

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/uri"
	"os"
	"os/user"
	"path"
	"strings"
)

const (
	REPO_URI_OVERWRITE_ENV_PREFIX = "OP_BUILD_URI="
)

// Get the uri overwrites from flags and environments
func getRepositoryUriOverwrites(flags []string) (map[string]*uri.RepositoryReference, error) {
	// Initialize the local path mapping by environment and add flags since we want to let flag overwrite the path from environment variables
	var paths []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, REPO_URI_OVERWRITE_ENV_PREFIX) {
			paths = append(paths, env[len(REPO_URI_OVERWRITE_ENV_PREFIX):])
		}
	}
	for _, flag := range flags {
		paths = append(paths, flag)
	}
	// Resolve
	return getRepositoryUriOverwritesFromLocalPathMapping(paths)
}

func getRepositoryUriOverwritesFromLocalPathMapping(paths []string) (map[string]*uri.RepositoryReference, error) {
	// Get repository path map from local path mapping list
	// The flag value looks like name:~/repository/...
	// The output is a map looks like { name: /home/user/repository/.. }
	//
	// Get the home path (to replace ~ in the path)
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	homeDir := u.HomeDir
	// Parse the paths
	repoPathMap := make(map[string]*uri.RepositoryReference)
	for _, repoPath := range paths {
		idx := strings.Index(repoPath, ":")
		if idx == -1 {
			return nil, errors.New(fmt.Sprint("Malformed repository path: ", repoPath))
		}
		name := repoPath[:idx]
		p := repoPath[idx+1:]
		if len(name) == 0 {
			return nil, errors.New(fmt.Sprint("Invalid repository path, require name: ", repoPath))
		}
		if len(p) == 0 {
			return nil, errors.New(fmt.Sprint("Invalid repository path, require path: ", repoPath))
		}
		// Check home
		if strings.HasPrefix(p, "~/") {
			p = path.Join(homeDir, p[2:])
		}
		// Set
		repoPathMap[name] = &uri.RepositoryReference{Local: p}
	}
	// Done
	return repoPathMap, nil
}

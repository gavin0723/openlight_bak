// Author: lipixun
// File Name: git.go
// Description:

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetGitRootPath returns the root path of git repository of given path
func GetGitRootPath(p string) (string, error) {
	cmd := exec.Command("git", "-C", p, "rev-parse", "--show-toplevel")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Git command failed: %v", err)
	}
	return strings.Trim(string(output), " \t\r\n"), nil
}

// GitRepositoryInfo defines the git repository info
type GitRepositoryInfo struct {
	Branch     string
	Commit     string
	CommitTime string
}

// GetGitRepositoryInfo returns the git repository info
func GetGitRepositoryInfo(p string) (*GitRepositoryInfo, error) {
	return nil, nil
}

// Author: lipixun
// File Name: git.go
// Description:

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
	// Run git command
	cmd := exec.Command("git", "show", "-s", "--format=%H,%ct")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Git command failed: %v", err)
	}
	parts := strings.Split(strings.Trim(string(output), " \t\r\n"), ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Malformed git show command output [%v]", string(output))
	}
	commit, commitTime := parts[0], parts[1]
	seconds, err := strconv.ParseInt(commitTime, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse commit time [%v]", commitTime)
	}
	commitTime = time.Unix(seconds, 0).Format(time.RFC3339)
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Git command failed: %v", err)
	}
	branch := strings.Trim(string(output), " \t\r\n")
	// Done
	return &GitRepositoryInfo{branch, commit, commitTime}, nil
}

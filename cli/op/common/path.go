// Author: lipixun
// File Name: path.go
// Description:

package common

import (
	"os"

	"github.com/ops-openlight/openlight/pkg/git"
)

// GetCurrentRepositoryPath returns the (git) repository root path by current path
func GetCurrentRepositoryPath() (string, error) {
	p, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return git.GetRootPath(p)
}

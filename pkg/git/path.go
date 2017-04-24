// Author: lipixun
// File Name: path.go
// Description:

package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetRootPath returns the root path of git repository of given path
func GetRootPath(p string) (string, error) {
	cmd := exec.Command("git", "-C", p, "rev-parse", "--show-toplevel")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Git command failed: %v", err)
	}
	return strings.Trim(string(output), " \t\r\n"), nil
}

// Author: lipixun
// Created Time : 五 10/21 16:06:00 2016
//
// File Name: clean.go
// Description:
//	Clean
package build

import (
	"github.com/ops-openlight/openlight"
	ocli "github.com/ops-openlight/openlight/cli"
	"github.com/ops-openlight/openlight/sourcecode/workspace"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func runClean(c *cli.Context) error {
	// Get verbose and logger
	verbose := c.GlobalBool("verbose")
	var level int
	if verbose {
		level = openlight.LogLevelAll
	} else {
		level = openlight.LogLevelInfo
	}
	logger := openlight.NewLogger(os.Stderr, level)
	// Run the clean on both current path and git root path
	p, err := os.Getwd()
	if err != nil {
		logger.WriteErrorHeaderln(CliLogHeader, "Failed to get current directory, error: ", err)
		return cli.NewExitError("", 1)
	}
	fs := workspace.NewUserWorkspaceFileSystem(p)
	fs.Clean()
	userPath, err := ocli.GetGitRootFromCurrentDirectory()
	if err != nil {
		fs := workspace.NewUserWorkspaceFileSystem(userPath)
		fs.Clean()
	}
	// Done
	return nil
}

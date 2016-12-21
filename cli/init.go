// Author: lipixun
// Created Time : 二 12/20 11:11:35 2016
//
// File Name: init.go
// Description:
//	The cli initialization
package cli

import (
	"fmt"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"gopkg.in/urfave/cli.v1"
)

// Get the workspace
func GetWorkspace(c *cli.Context) (*workspace.Workspace, error) {
	verbose := c.GlobalBool("verbose")
	workDirProjectPath := c.GlobalString("workdir-project-path")
	workDirUserPath := c.GlobalString("workdir-user-path")
	workDirGlobalPath := c.GlobalString("workdir-global-path")
	dockerUri := c.GlobalString("docker-uri")
	// Create workspace options
	options := workspace.NewWorkspaceOptions()
	options.Verbose = verbose
	options.EnableColor = true
	options.Dir.GlobalPath = workDirGlobalPath
	options.Dir.UserPath = workDirUserPath
	if workDirProjectPath == "" {
		options.Dir.CurrentPathAsProjectPath = true
	} else {
		options.Dir.CurrentPathAsProjectPath = false
		options.Dir.ProjectPath = workDirProjectPath
	}
	options.ThirdService.Docker.Uri = dockerUri
	// Create workspace
	ws, err := workspace.New(options, nil)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to initialize workspace, error: %s", err), 1)
	} else {
		// Done
		return ws, nil
	}
}

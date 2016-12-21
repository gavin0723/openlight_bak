// Author: lipixun
// Created Time : 二 10/18 21:51:04 2016
//
// File Name: main.go
// Description:
//	The openlight cli main entry
package main

import (
	"fmt"
	"github.com/ops-openlight/openlight/cli/build"
	"github.com/ops-openlight/openlight/cli/runner"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"gopkg.in/urfave/cli.v1"
	"os"
)

var (
	buildBranch string
	buildCommit string
	buildTime   string
	buildTag    string
	// The version string
	Version = fmt.Sprintf("Branch [%s] Commit [%s] Build Time [%s] Tag [%s]", buildBranch, buildCommit, buildTime, buildTag)
)

func main() {
	// The main entry
	// Create cli application
	app := cli.NewApp()
	app.Name = "op"
	app.Usage = "Openlight CLI"
	app.Version = Version
	// Global flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Show verbose log (debug log)",
		},
		cli.StringFlag{
			Name:  "workdir-project-path",
			Usage: "The openlight project workdir path",
		},
		cli.StringFlag{
			Name:  "workdir-user-path",
			Value: workspace.DefaultUserDirPath,
			Usage: "The openlight user workdir path",
		},
		cli.StringFlag{
			Name:  "workdir-global-path",
			Value: workspace.DefaultGlobalDirPath,
			Usage: "The openlight global workdir path",
		},
		cli.StringFlag{
			Name:  "docker-uri",
			Value: workspace.DefaultDockerServiceUri,
			Usage: "The docker daemon uri",
		},
	}
	// Add commands from modules
	for _, cmd := range build.GetCommand() {
		app.Commands = append(app.Commands, cmd)
	}
	for _, cmd := range runner.GetCommand() {
		app.Commands = append(app.Commands, cmd)
	}
	// Run it
	app.Run(os.Args)
}

// Author: lipixun
// Created Time : 二 10/18 21:51:04 2016
//
// File Name: main.go
// Description:
//	The openlight cli main entry
package main

import (
	"fmt"
	"os"

	"github.com/ops-openlight/openlight/cli/op/build"
	"github.com/ops-openlight/openlight/cli/op/runner"
	"github.com/ops-openlight/openlight/cli/op/spec"
	"gopkg.in/urfave/cli.v1"
)

var (
	buildBranch     string
	buildCommit     string
	buildCommitTime string
	buildTime       string
	buildTag        string
	// Version The version string
	Version = fmt.Sprintf("Branch [%v] Commit [%v] Commit Time [%v] Build Time [%v] Tag [%v]", buildBranch, buildCommit, buildCommitTime, buildTime, buildTag)
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
	}
	// Add commands from modules
	app.Commands = append(app.Commands, spec.GetCommands()...)
	app.Commands = append(app.Commands, build.GetCommands()...)
	app.Commands = append(app.Commands, runner.GetCommands()...)
	// Run it
	app.Run(os.Args)
}

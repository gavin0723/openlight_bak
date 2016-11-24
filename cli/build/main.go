// Author: lipixun
// Created Time : 四 10/20 10:04:23 2016
//
// File Name: main.go
// Description:
//
package build

import (
	"gopkg.in/urfave/cli.v1"
)

const (
	CliLogHeader = "Openlight CLI"
)

func GetCommand() []cli.Command {
	return []cli.Command{
		{
			Category: "Builder",
			Name:     "local-build",
			Aliases:  []string{"lb"},
			Usage:    "Force build with local dependencies. The same as 'op build --only-local' ",
			Action:   runLocalBuild,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "repository-path, p",
					Usage: "Define the local repository path. Format: name:path",
				},
			},
		},
		{
			Category: "Builder",
			Name:     "clean-build",
			Usage:    "Clean the workspace",
			Action:   runClean,
		},
	}
}

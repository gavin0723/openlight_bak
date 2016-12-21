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
	LogHeader = "CLI.Builder"
)

func GetCommand() []cli.Command {
	return []cli.Command{
		{
			Category: "Builder",
			Name:     "local-build",
			Aliases:  []string{"lb"},
			Usage:    "Force build with local dependencies. The same as 'op build --only-local' ",
			Action:   LocalBuild,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Value: "build",
					Usage: "The output path",
				},
				cli.BoolFlag{
					Name:  "disable-finder",
					Usage: "Disable the repository local finder",
				},
				cli.StringSliceFlag{
					Name:  "repository-remote-overwrite, w",
					Usage: "Overwrite the repository remote (or local path). Format: uri:path",
				},
			},
		},
		{
			Category: "Builder",
			Name:     "clean-build",
			Usage:    "Clean the build workspace. This will clean user ALL build data",
			Action:   Clean,
		},
	}
}

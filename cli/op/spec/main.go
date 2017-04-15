// Author: lipixun
// Created Time : 一  3/13 21:03:44 2017
//
// File Name: main.go
// Description:

package spec

import (
	"gopkg.in/urfave/cli.v1"
)

// GetCommands Get build commands
func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "spec",
			Usage: "Openlight Spec File Utility",
			Subcommands: []cli.Command{
				{
					Name:  "parse",
					Usage: "Parse spec file(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file,f",
							Usage: "The spec (root) file. Will use stdin if not specified.",
						},
						cli.StringFlag{
							Name:  "output,o",
							Usage: "Write the parsed spec data model to file.",
						},
					},
					Action: parseSpec,
				},
			},
		},
	}
}

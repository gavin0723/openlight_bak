// Author: lipixun
// Created Time : 一  3/13 21:03:44 2017
//
// File Name: main.go
// Description:

package build

import (
	"gopkg.in/urfave/cli.v1"
)

// GetCommands Get build commands
func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "build",
			Usage: "Build packages and targets",
		},
	}
}

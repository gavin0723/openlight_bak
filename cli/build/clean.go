// Author: lipixun
// Created Time : 五 10/21 16:06:00 2016
//
// File Name: clean.go
// Description:
//	Clean
package build

import (
	opcli "github.com/ops-openlight/openlight/cli"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode/builder"
	"gopkg.in/urfave/cli.v1"
)

func Clean(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Run clean
	if err := builder.CleanBuildData(ws); err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to clean build data, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Done
	return nil
}

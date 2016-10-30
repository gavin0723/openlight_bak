// Author: lipixun
// Created Time : 五 10/21 16:05:51 2016
//
// File Name: localbuild.go
// Description:
//	Local build
package build

import (
	"fmt"
	"github.com/ops-openlight/openlight"
	ocli "github.com/ops-openlight/openlight/cli"
	"github.com/ops-openlight/openlight/sourcecode/workspace"
	"github.com/ops-openlight/openlight/uri"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func runLocalBuild(c *cli.Context) error {
	// Get verbose and logger
	verbose := c.GlobalBool("verbose")
	var level int
	if verbose {
		level = openlight.LogLevelAll
	} else {
		level = openlight.LogLevelInfo
	}
	logger := openlight.NewLogger(os.Stderr, level)
	// Get build target uris
	var targetUris []*uri.TargetUri
	for _, targetUriArg := range c.Args() {
		targetUri := uri.ParseTargetUri(targetUriArg)
		if targetUri == nil {
			logger.WriteErrorHeaderln(CliLogHeader, "Failed to parse target uri from arg", targetUriArg)
			return cli.NewExitError("", 1)
		}
		targetUris = append(targetUris, targetUri)
	}
	// Get repository uri overwrites
	repoUriOverwrites, err := getRepositoryUriOverwrites(c.StringSlice("repository-path"))
	if err != nil {
		logger.WriteErrorHeaderln(CliLogHeader, fmt.Sprint("Failed to load repository uri overwrites, error: ", err))
		return cli.NewExitError("", 1)
	}
	// Adjust the target uri, create workspace file system
	userPath, err := ocli.GetGitRootFromCurrentDirectory()
	if err != nil {
		// Failed to get git root, check the target uris
		if len(targetUris) == 0 {
			logger.WriteErrorHeaderln(CliLogHeader, "Failed to get current git root directory (and which is required by empty target uris), error: ", err)
			return cli.NewExitError("", 1)
		}
		for _, targetUri := range targetUris {
			if targetUri.Repository.Local == "" && targetUri.Repository.Remote == nil {
				logger.WriteErrorHeaderln(CliLogHeader, "Failed to get current git root directory (and which is required by target", targetUri.Target, "), error: ", err)
				return cli.NewExitError("", 1)
			}
		}
	} else {
		// Great, check the target uris
		if len(targetUris) == 0 {
			// No target uri defined, add current repository
			targetUris = append(targetUris, &uri.TargetUri{Repository: uri.RepositoryReference{Local: userPath}})
		} else {
			for _, targetUri := range targetUris {
				if targetUri.Repository.Local == "" && targetUri.Repository.Remote == nil {
					// Set as local
					targetUri.Repository.Local = userPath
				}
			}
		}
	}
	fs := workspace.NewUserWorkspaceFileSystem(userPath)
	if err := fs.Initialize(); err != nil {
		logger.WriteErrorHeaderln(CliLogHeader, "Failed to initialize workspace filesystem, error: ", err)
		return cli.NewExitError("", 1)
	}
	// Create workspace
	wsOptions := workspace.LocalWorkspaceOptions()
	wsOptions.Verbose = verbose
	ws, err := workspace.NewWorkspace(fs, wsOptions, logger)
	if repoUriOverwrites != nil {
		ws.Repository.UriOverwrites = repoUriOverwrites
	}
	// Start build
	return startBuild(targetUris, ws)
}

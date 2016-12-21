// Author: lipixun
// Created Time : 六 10/29 19:43:43 2016
//
// File Name: build.go
// Description:
//	The build
package build

import (
	"errors"
	"fmt"
	opcli "github.com/ops-openlight/openlight/cli"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/sourcecode/builder"
	"github.com/ops-openlight/openlight/pkg/sourcecode/graph"
	"github.com/ops-openlight/openlight/pkg/sourcecode/repoloader"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"github.com/ops-openlight/openlight/pkg/uri"
	"github.com/ops-openlight/openlight/pkg/util"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"strings"
)

const (
	REPO_URI_OVERWRITE_ENV_PREFIX = "OP_SOURCECODE_REPO_"
)

// Local build command
func LocalBuild(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Get build target uris
	var targetUris []*uri.TargetUri
	for _, targetUriArg := range c.Args() {
		targetUri := uri.ParseTargetUri(targetUriArg)
		if targetUri == nil {
			logger.LeveledPrintf(log.LevelError, "Failed to parse target uri from arg: %s\n", targetUriArg)
			return cli.NewExitError("", 1)
		}
		targetUris = append(targetUris, targetUri)
	}
	// Get options
	disableFinder := c.Bool("disable-finder")
	// Get repository uri overwrites
	remoteOverwrites, err := getRemoteOverwrites(c.StringSlice("repository-remote-overwrite"), logger)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to load repository remote overwrites, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	if ws.Verbose {
		showRemoteOverwrites(remoteOverwrites, ws.Logger)
	}
	// Get the output path
	realPath, err := util.GetRealPath(c.String("output"))
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to get output real path, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	output, err := filepath.Abs(realPath)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to get output abs path, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Adjust the target uri, create workspace file system
	currentProjectRootPath, err := opcli.GetGitRootFromCurrentDirectory()
	if err != nil {
		// Failed to get git root, check the target uris
		if len(targetUris) == 0 {
			logger.LeveledPrintf(log.LevelError, "Failed to get current git root directory (and which is required by empty target uris), error: %s\n", err)
			return cli.NewExitError("", 1)
		}
		for _, targetUri := range targetUris {
			if targetUri.Repository != nil {
				logger.LeveledPrintf(log.LevelError, "Failed to get current git root directory (and which is required by target %s), error: %s\n", targetUri.Name, err)
				return cli.NewExitError("", 1)
			}
		}
	} else {
		// Great, check the target uris
		if len(targetUris) == 0 {
			// No target uri defined, add current repository
			target, err := getDefaultTargetUri(currentProjectRootPath)
			if err != nil {
				logger.LeveledPrintf(log.LevelError, "Failed to current target uri, error: %s\n", err)
				return cli.NewExitError("", 1)
			}
			targetUris = append(targetUris, target)
		} else {
			for _, targetUri := range targetUris {
				if targetUri.Repository == nil {
					targetUri.Repository = &uri.RepositoryUri{Uri: currentProjectRootPath}
				}
			}
		}
	}
	// Start build
	options := BuildOptions{
		AllowLocal:       true,
		OnlyLocal:        true,
		Output:           output,
		DisableFinder:    disableFinder,
		RemoteOverwrites: remoteOverwrites,
	}
	return build(targetUris, ws, options, logger)
}

func Build(c *cli.Context) error {
	return cli.NewExitError("Not implemented", 1)
}

// Get the uri overwrites from flags and environments
func getRemoteOverwrites(flags []string, logger log.Logger) (map[string]string, error) {
	// Initialize the local path mapping by environment and add flags since we want to let flag overwrite the path from environment variables
	var remoteOverwrites map[string]string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, REPO_URI_OVERWRITE_ENV_PREFIX) {
			idx := strings.Index(env, "=")
			if idx != -1 {
				uri := env[len(REPO_URI_OVERWRITE_ENV_PREFIX):idx]
				if len(uri) == 0 {
					logger.LeveledPrintf(log.LevelWarn, "Failed to load repository remote overwrites from environment variable [%s], err: Invalid uri\n", env)
					continue
				}
				path, err := util.GetRealPath(env[idx+1:])
				if err != nil {
					logger.LeveledPrintf(log.LevelWarn, "Failed to load repository remote overwrites from environment variable [%s], err: %s\n", env, err)
					continue
				}
				// Add it
				remoteOverwrites[uri] = path
			}
		}
	}
	for _, flag := range flags {
		idx := strings.Index(flag, ":")
		if idx == -1 {
			return nil, errors.New(fmt.Sprintf("Malformed repository remote overwrites argument [%s], : not found", flag))
		}
		uri := flag[:idx]
		if len(uri) == 0 {
			return nil, errors.New(fmt.Sprintf("Malformed repository remote overwrites argument [%s], Invalid uri", flag))
		}
		path, err := util.GetRealPath(flag[idx+1:])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Malformed repository remote overwrites argument [%s], error: %s", err))
		}
		remoteOverwrites[uri] = path
	}
	// Done
	return remoteOverwrites, nil
}

func showRemoteOverwrites(remoteOverwrites map[string]string, logger log.Logger) {
	for uri, remote := range remoteOverwrites {
		logger.LeveledPrintf(log.LevelDebug, "Use repository remote overwrite %s --> %s\n", uri, remote)
	}
}

// Get the default target uri
func getDefaultTargetUri(path string) (*uri.TargetUri, error) {
	spec, err := repoloader.LoadRepositorySpecFromFile(filepath.Join(path, spec.SpecFileName))
	if err != nil {
		return nil, err
	}
	name := spec.Options.Default.Build.Target
	if name == "" {
		return nil, errors.New("No default target defined")
	}
	return &uri.TargetUri{Name: name, Repository: &uri.RepositoryUri{Uri: path}}, nil
}

type BuildOptions struct {
	AllowLocal       bool
	OnlyLocal        bool
	Output           string
	DisableFinder    bool
	RemoteOverwrites map[string]string
}

// Start the build process
func build(targetUris []*uri.TargetUri, ws *workspace.Workspace, options BuildOptions, logger log.Logger) error {
	// Load the source code graph
	g, err := graph.New(ws, graph.GraphOptions{UseLocalDependency: options.AllowLocal, DisableFinder: options.DisableFinder})
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create sourcecode graph, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	if len(options.RemoteOverwrites) > 0 {
		for uri, remote := range options.RemoteOverwrites {
			g.RemoteOverwrites[uri] = remote
		}
	}
	// Load the repository with the targets
	var targets []*spec.Target
	for _, targetUri := range targetUris {
		r, err := g.Load(targetUri.Repository.Uri, graph.LoadOptions{Branch: targetUri.Repository.Branch, Commit: targetUri.Repository.Commit, Targets: []string{targetUri.Name}})
		if err != nil {
			logger.LeveledPrintf(log.LevelError, "Failed to load target [%s] remote [%s], err: %s\n", targetUri.Name, targetUri.Repository.Uri, err)
			return cli.NewExitError("", 1)
		}
		target := g.Targets[spec.GetTargetKey(targetUri.Name, r)]
		if target == nil {
			logger.LeveledPrintf(log.LevelError, "Target [%s] not loaded after repository loaded\n", targetUri.Name)
			return cli.NewExitError("", 1)
		}
		targets = append(targets, target)
	}
	// Create the builder
	buildTag, err := builder.NewTag()
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to generate build tag, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	logger.LeveledPrintf(log.LevelWarn, "Build tag generated: %s\n", buildTag)
	b, err := builder.New(g, builder.NewBuilderOptions(buildTag, options.Output))
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create builder, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Build the targets
	for _, target := range targets {
		logger.Printf("Start build target %s\n", target.Key())
		buildResult, err := b.Build(target)
		if err != nil {
			logger.LeveledPrintf(log.LevelError, "Failed to build target [%s] error: %s\n", target.Key(), err)
			return cli.NewExitError("", 1)
		}
		for name, art := range buildResult.Artifacts {
			logger.Printf("\tArtifact generated: %s --> %s\n", name, art.String())
		}
	}
	logger.Println("Build completed")
	// Done
	return nil
}

// Author: lipixun
// Created Time : 六 10/29 19:43:43 2016
//
// File Name: build.go
// Description:
//	The build
package build

import (
	"github.com/ops-openlight/openlight/sourcecode/target"
	"github.com/ops-openlight/openlight/sourcecode/workspace"
	"github.com/ops-openlight/openlight/uri"
	"gopkg.in/urfave/cli.v1"
)

// Start build
func startBuild(targetUris []*uri.TargetUri, ws *workspace.Workspace) error {
	var targets []*target.Target
	// Construct the target graph
	graph := target.NewTargetGraph()
	for _, targetUri := range targetUris {
		// Load repository
		repo, err := ws.Repository.Load("", &targetUri.Repository, ws.DefaultRepositoryPathFunction)
		if err != nil {
			ws.Logger.WriteErrorHeaderln(CliLogHeader, "Failed to load repository [", targetUri.Repository.String(), "] , error: ", err)
			return cli.NewExitError("", 1)
		}
		if ws.Verbose() {
			ws.Logger.WriteDebugHeaderln(CliLogHeader, "Loaded repository [", repo.Uri(), "] with metadata: ", repo.Metadata.String())
		}
		// Load the target
		target, err := graph.Load(targetUri.Target, repo, ws)
		if err != nil {
			ws.Logger.WriteErrorHeaderln(CliLogHeader, "Failed to load target [", targetUri.Target, "] , error: ", err)
			return cli.NewExitError("", 1)
		}
		// Good
		targets = append(targets, target)
	}
	// Generate a build tag
	buildTag, err := uri.NewTag()
	if err != nil {
		ws.Logger.WriteErrorHeaderln(CliLogHeader, "Failed to generate build tag, error: ", err)
		return cli.NewExitError(err.Error(), 1)
	}
	ws.Logger.WriteInfoHeaderln(CliLogHeader, "Build with tag ", buildTag)
	// Build the targets
	builder, err := graph.NewBuilder(target.NewBuildOption(buildTag), ws)
	if err != nil {
		ws.Logger.WriteErrorHeaderln(CliLogHeader, "Failed to create graph builder, error: ", err)
		return cli.NewExitError("", 1)
	}
	for _, target := range targets {
		ws.Logger.WriteInfoHeaderln(CliLogHeader, "Start build target ", target.Key())
		buildResult, err := builder.Build(target)
		if err != nil {
			ws.Logger.WriteErrorHeaderln(CliLogHeader, "Failed to build target ", target.Key(), " , error: ", err)
			return cli.NewExitError("", 1)
		}
		// Print out the build result
		if buildResult != nil && len(buildResult.Artifacts) > 0 {
			ws.Logger.WriteInfoHeaderln(CliLogHeader, "Artifact generated")
			for _, artifact := range buildResult.Artifacts {
				ws.Logger.WriteInfoHeaderln(CliLogHeader, "\t", artifact.String())
			}
		}
	}
	ws.Logger.WriteInfoHeaderln(CliLogHeader, "Build completed")
	// Done
	return nil
}

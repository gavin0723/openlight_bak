// Author: lipixun
// Created Time : 一  3/13 21:02:14 2017
//
// File Name: spec.go
// Description:

package spec

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ops-openlight/openlight/pkg/rule"
	"github.com/ops-openlight/openlight/pkg/rule/modules/build"
	"github.com/ops-openlight/openlight/pkg/rule/modules/runner"
)

func parseSpec(c *cli.Context) error {
	filename, outputFilename := c.String("file"), c.String("output")
	// Load filename
	engine := rule.NewEngine()
	ctx, err := engine.ParseFile(filename)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to parse file, error: %v", err), 1)
	}
	// Open output
	var out io.Writer
	if outputFilename != "" {
		outFile, err := os.Create(outputFilename)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to create file, error: %v", err), 1)
		}
		defer outFile.Close()
		out = outFile
	} else {
		out = os.Stdout
	}
	jMarshaler := jsonpb.Marshaler{}
	// Write build module
	buildfile := ctx.GetModule("build").(build.Module).Spec()
	if buildfile != nil {
		if err := jMarshaler.Marshal(out, buildfile); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to dump BuildFile, error: %v", err), 1)
		}
		fmt.Fprintln(out)
	}
	// Write runner module
	runfile := ctx.GetModule("runner").(runner.Module).Spec()
	if runfile != nil {
		if err := jMarshaler.Marshal(out, runfile); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to dump RunnerFile, error: %v", err), 1)
		}
		fmt.Fprintln(out)

	}
	// Done
	return nil
}

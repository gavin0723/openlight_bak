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
	loader, err := rule.NewFileLoader([]string{filename})
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
	// Write output
	buildModule := loader.GetModule("build").(build.Module)
	jMarshaler := jsonpb.Marshaler{}
	for _, pkg := range buildModule.Packages() {
		pbPackage, err := pkg.GetProto()
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get protobuf object of package [%v], error: %s", pkg.Name, err), 1)
		}
		if err := jMarshaler.Marshal(out, pbPackage); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to dump protobuf object of package [%v], error: %s", pkg.Name, err), 1)
		}
		fmt.Fprintln(out)
	}
	runnerModule := loader.GetModule("runner").(runner.Module)
	for _, cmd := range runnerModule.Commands() {
		pbCommand, err := cmd.GetProto()
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get protobuf object of command [%v], error: %s", cmd.ID, err), 1)
		}
		if err := jMarshaler.Marshal(out, pbCommand); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to dump protobuf object of command [%v], error: %s", cmd.ID, err), 1)
		}
		fmt.Fprintln(out)
	}
	// Done
	return nil
}

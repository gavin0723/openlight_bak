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
	// Write the rule files
	ruleFiles, err := ctx.GetRule()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if err := jMarshaler.Marshal(out, ruleFiles); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to dump RuleFiles, error: %v", err), 1)
	}
	fmt.Fprintln(out)
	// Done
	return nil
}

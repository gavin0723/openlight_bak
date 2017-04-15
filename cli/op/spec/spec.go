// Author: lipixun
// Created Time : 一  3/13 21:02:14 2017
//
// File Name: spec.go
// Description:

package spec

import (
	"fmt"

	"gopkg.in/urfave/cli.v1"

	"github.com/ops-openlight/openlight/pkg/rule"
	"github.com/ops-openlight/openlight/pkg/rule/modules/build"
)

func parseSpec(c *cli.Context) error {
	filename := c.String("file")
	// Load filename
	loader, err := rule.NewFileLoader([]string{filename})
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to parse file, error: %v", err), 1)
	}
	m := loader.GetModule("build").(build.Module)
	for _, pkg := range m.Packages() {
		fmt.Println(pkg.Name)
	}
	// Done
	return nil
}

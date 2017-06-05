// Author: lipixun
//
// File Name: main.go
// Description:

package runner

import (
	"gopkg.in/urfave/cli.v1"
)

// GetCommands gets runner commands
func GetCommands() []cli.Command {
	var runner Runner
	return runner.GetCommands()
}

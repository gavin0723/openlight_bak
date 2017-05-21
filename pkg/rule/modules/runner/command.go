// Author: lipixun
// File Name: command.go
// Description:

package runner

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/rule/common"
)

// Ensure the interface is implemented
var _ common.Object = (*RunCommand)(nil)

// Exposed lua infos
const (
	runCommandLUAName     = "Command"
	runCommandLUATypeName = "runner-command"
)

// RunCommand implements RunCommand for lua
type RunCommand pbSpec.RunCommand

// NewRunCommand creates a new RunCommand
func NewRunCommand(name, comment string, args []string, workdir string, envs []string) *RunCommand {
	return &RunCommand{
		Name:    name,
		Comment: comment,
		Args:    args,
		Workdir: workdir,
		Envs:    envs,
	}
}

// NewRunCommandFromLUA creates a new RunCommand from lua
func NewRunCommandFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	name, err := params.GetString("name")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [name]: %v", err)
	}
	if name == "" {
		return nil, errors.New("Require parameter [name]")
	}
	args, err := params.GetStringSlice("args")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [args]: %v", err)
	}
	comment, err := params.GetString("comment")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [comment]: %v", err)
	}
	workdir, err := params.GetString("workdir")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [workdir]: %v", err)
	}
	envs, err := params.GetStringSlice("envs")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [envs]: %v", err)
	}
	// Create command
	cmd := NewRunCommand(name, comment, args, workdir, envs)
	// Done
	return cmd.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (cmd *RunCommand) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = cmd
	L.SetMetatable(ud, L.GetTypeMetatable(runCommandLUATypeName))
	// Done
	return ud
}

// registerRunCommandType registers RunCommand type in lua
func registerRunCommandType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(runCommandLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewRunCommandFromLUA))
	// Add type to module
	L.SetField(mod, runCommandLUAName, mt)
}

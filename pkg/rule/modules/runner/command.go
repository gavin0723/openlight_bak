// Author: lipixun
// File Name: command.go
// Description:

package runner

import (
	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
	lua "github.com/yuin/gopher-lua"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

// Exposed lua infos
const (
	LUANameRunCommand = "Command"
	LUATypeRunCommand = "Runner-Command"
)

// RunCommand represents the runner command
type RunCommand struct {
	LUA.Object
	ID   string
	Name string
	Args []string
}

// NewRunCommand creates a new RunCommand
func NewRunCommand(id, name string, args []string, options *lua.LTable) *RunCommand {
	var cmd = RunCommand{
		ID:   id,
		Name: name,
		Args: args,
	}
	cmd.Object = LUA.NewObject(LUATypeRunCommand, options, &cmd)
	// Done
	return &cmd
}

// GetProto returns the protobuf object
func (cmd *RunCommand) GetProto() (*pbSpec.RunCommand, error) {
	var pbRunCommand pbSpec.RunCommand
	pbRunCommand.Id = cmd.ID
	pbRunCommand.Name = cmd.Name
	pbRunCommand.Args = cmd.Args
	// Get options
	var err error
	options := cmd.GetOptions()
	if pbRunCommand.Workdir, err = LUA.TryGetStringFromTable(options, "workdir", ""); err != nil {
		return nil, err
	}
	if pbRunCommand.Envs, err = LUA.TryGetStringSliceFromTable(options, "envs", nil); err != nil {
		return nil, err
	}
	// Done
	return &pbRunCommand, nil
}

// RunCommandLUAFuncs defines all lua functions for RunCommand
var RunCommandLUAFuncs = map[string]lua.LGFunction{
	"name":    LUAFuncRunCommandName,
	"options": LUA.FuncObjectOptions,
}

// RegisterRunCommandType registers RunCommand type in lua
func RegisterRunCommandType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(LUATypeRunCommand)
	L.SetField(mt, "new", L.NewFunction(LUARunCommandNew))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), RunCommandLUAFuncs))
	// Add type to module
	L.SetField(mod, LUANameRunCommand, mt)
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUARunCommandNew defines RunCommand.new function in lua
func LUARunCommandNew(L *lua.LState) int {
	id := L.CheckString(1)
	if id == "" {
		L.ArgError(1, "Require id")
		return 0
	}
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	args, err := LUA.ConvertTableToStringSlice(L.CheckTable(3))
	if err != nil {
		L.ArgError(3, "Require arguments list")
		return 0
	}
	// Create
	cmd := NewRunCommand(id, name, args, L.ToTable(4))
	// Return
	L.Push(cmd.GetLUAUserData(L))
	return 1
}

// LUARunCommandSelf get lua RunCommand self
func LUARunCommandSelf(L *lua.LState) *RunCommand {
	ud := L.CheckUserData(1)
	if pkg, ok := ud.Value.(*RunCommand); ok {
		return pkg
	}
	L.ArgError(1, "RunCommand expected")
	return nil
}

// LUAFuncRunCommandName defines RunCommand.id in lua
func LUAFuncRunCommandName(L *lua.LState) int {
	cmd := LUARunCommandSelf(L)
	if cmd == nil {
		return 0
	}
	if L.GetTop() == 1 {
		L.Push(lua.LString(cmd.Name))
		return 1
	} else if L.GetTop() == 2 {
		name := L.CheckString(2)
		if name == "" {
			L.ArgError(2, "Require name")
			return 0
		}
		cmd.Name = name
		return 0
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

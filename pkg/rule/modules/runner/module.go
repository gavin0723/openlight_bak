// Author: lipixun
// File Name: module.go
// Description:

package runner

import (
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

const (
	// LUANameModule defines the name of BuildModule
	LUANameModule = "runner"
)

// Ensure the interface is implements
var _ Module = (*_Module)(nil)

// RegisterTypes registers all types in this module
func RegisterTypes(L *lua.LState, mod *lua.LTable) {
	RegisterRunCommandType(L, mod)
}

// Module defines the module used by build progress
type Module interface {
	LUA.Module
	// Commands returns all commands
	Commands() []*RunCommand
	// CommandSpec returns all commands in proto format
	CommandSpec() ([]*pbSpec.RunCommand, error)
}

// _Module implements the Module interface
type _Module struct {
	ctx      LUA.ModuleContext
	commands map[string]*RunCommand
}

// NewModule returns new Module
func NewModule(ctx LUA.ModuleContext) Module {
	return &_Module{
		ctx:      ctx,
		commands: make(map[string]*RunCommand),
	}
}

// InitLInitLUAModule initializes (preload) module into lua
func (m *_Module) InitLUAModule(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"getCommand":    m.LUAFuncGetCommand,
		"addCommand":    m.LUAFuncAddCommand,
		"deleteCommand": m.LUAFuncDeleteCommand,
	})
	RegisterTypes(L, mod)
	L.Push(mod)
	// Done
	return 1
}

// Name returns the module name
func (m *_Module) Name() string {
	return LUANameModule
}

// Commands returns all commands
func (m *_Module) Commands() []*RunCommand {
	var commands []*RunCommand
	for _, cmd := range m.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// CommandSpec returns all commands in proto format
func (m *_Module) CommandSpec() ([]*pbSpec.RunCommand, error) {
	var pbCommands []*pbSpec.RunCommand
	for _, cmd := range m.commands {
		pbCommand, err := cmd.GetProto()
		if err != nil {
			return nil, err
		}
		pbCommands = append(pbCommands, pbCommand)
	}
	return pbCommands, nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUAFuncGetCommand defines runner.getCommand in lua
func (m *_Module) LUAFuncGetCommand(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.ArgError(0, "Invalid arguments")
		return 0
	}
	cmd := m.commands[L.CheckString(1)]
	if cmd != nil {
		L.Push(cmd.GetLUAUserData(L))
		return 1
	}
	// Not found
	return 0
}

// LUAFuncAddCommand defines runner.addCommand in lua
func (m *_Module) LUAFuncAddCommand(L *lua.LState) int {
	for i := 1; i <= L.GetTop(); i++ {
		ud := L.CheckUserData(i)
		if ud == nil {
			return 0
		}
		cmd, ok := ud.Value.(*RunCommand)
		if !ok {
			L.ArgError(i, "Not a command")
			return 0
		}
		if _cmd := m.commands[cmd.ID]; _cmd != nil {
			L.ArgError(i, fmt.Sprintf("Duplicated command [%v]", _cmd.ID))
			return 0
		}
		m.commands[cmd.ID] = cmd
	}
	// Not found
	return 0
}

// LUAFuncDeleteCommand defines runner.deleteCommand in lua
func (m *_Module) LUAFuncDeleteCommand(L *lua.LState) int {
	for i := 1; i < L.GetTop(); i++ {
		delete(m.commands, L.CheckString(i))
	}
	// Done
	return 0
}

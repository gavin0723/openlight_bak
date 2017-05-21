// Author: lipixun
// File Name: module.go
// Description:

package runner

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/rule/common"
	"github.com/ops-openlight/openlight/pkg/rule/engine"
)

const (
	// ModuleName defines the module name
	ModuleName = "runner"
)

const (
	moduleLUAName = "runner"
)

// Ensure the interface is implements
var _ Module = (*_Module)(nil)
var _ engine.ModuleFactory = (*ModuleFactory)(nil)

// ModuleFactory implements engine.ModuleFactory to runner module
type ModuleFactory struct{}

// NewModuleFactory creates a new ModuleFactory
func NewModuleFactory() engine.ModuleFactory {
	return new(ModuleFactory)
}

// Name returns the name of the created module (which can be used for other modules to get this one)
func (factory *ModuleFactory) Name() string {
	return ModuleName
}

// Create a new module
func (factory *ModuleFactory) Create(ctx *engine.Context) (engine.Module, error) {
	if ctx == nil {
		return nil, errors.New("Require context")
	}
	return &_Module{ctx: ctx}, nil
}

// Module defines the runner model interface
type Module interface {
	engine.Module
	// Spec returns the run file object
	Spec() *pbSpec.RunFile
}

// _Module implements the Module interface
type _Module struct {
	ctx      *engine.Context
	commands *common.NamedCollection
}

// LUAName returns the module name in lua
func (m *_Module) LUAName() string {
	return moduleLUAName
}

// InitLUAModule initializes module into lua
func (m *_Module) InitLUAModule(L *lua.LState) int {
	// Create new module
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"commands": m.luaFuncCommands,
	})
	// Register types
	registerRunCommandType(L, mod)
	// Return module
	L.Push(mod)
	return 1
}

// Commands returns all commands
func (m *_Module) Spec() *pbSpec.RunFile {
	if m.commands == nil {
		return nil
	}
	var runfile pbSpec.RunFile
	runfile.Commands = make(map[string]*pbSpec.RunCommand)
	for _, item := range m.commands.Items() {
		runfile.Commands[item.Name] = (*pbSpec.RunCommand)(item.Value.(*RunCommand))
	}
	// Done
	return &runfile
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// luaFuncCommands defines runner.commands in lua
func (m *_Module) luaFuncCommands(L *lua.LState) int {
	if L.GetTop() != 0 {
		L.ArgError(0, "Invalid arguments")
	}
	// Return commands
	if m.commands == nil {
		m.commands = common.NewNamedCollection(m.convertCommand)
	}
	L.Push(m.commands.GetLUAUserData(L))
	return 1
}

func (m *_Module) convertCommand(value lua.LValue) (common.Object, error) {
	if value == nil {
		return nil, errors.New("Invalid nil value")
	}
	if value.Type() != lua.LTUserData {
		return nil, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTUserData, value.Type())
	}
	cmd, ok := value.(*lua.LUserData).Value.(*RunCommand)
	if !ok {
		return nil, errors.New("Expect RunCommand object")
	}
	return cmd, nil
}

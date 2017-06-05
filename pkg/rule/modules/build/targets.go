// Author: lipixun
// File Name: targets.go
// Description:

package build

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/rule/common"
)

// Ensure the interface is implemented
var _ Target = (*GeneralTarget)(nil)
var _ Target = (*CommandTarget)(nil)
var _ Target = (*GoBinaryTarget)(nil)
var _ Target = (*PythonLibTarget)(nil)

const (
	targetLUAName              = "Target"
	targetLUATypeName          = "build-target"
	commandTargetLUAName       = "CommandTarget"
	commandTargetLUATypeName   = "build-target-command"
	goBinaryTargetLUAName      = "GoBinaryTarget"
	goBinaryTargetLUATypeName  = "build-target-gobinary"
	pythonLibTargetLUAName     = "PythonLibTarget"
	pythonLibTargetLUATypeName = "build-target-pythonlib"
)

// registerTargetType registers TargetType type in lua
func registerTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(targetLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewGeneralTargetFromLUA))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"description": luaFuncTargetDescription,
		"dependent":   luaFuncTargetDependent,
	}))
	L.SetField(mod, targetLUAName, mt)
}

// registerCommandTargetType registers CommandTargetType type in lua
func registerCommandTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(commandTargetLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewCommandTargetFromLUA))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"description": luaFuncTargetDescription,
		"dependent":   luaFuncTargetDependent,
	}))
	L.SetField(mod, commandTargetLUAName, mt)
}

// registerGoBinaryTargetType registers GoBinaryTarget type in lua
func registerGoBinaryTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(goBinaryTargetLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewGoBinaryTargetFromLUA))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"description": luaFuncTargetDescription,
		"dependent":   luaFuncTargetDependent,
	}))
	L.SetField(mod, goBinaryTargetLUAName, mt)
}

// registerPythonLibTargetType registers PythonLibTarget type in lua
func registerPythonLibTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(pythonLibTargetLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewPythonLibTargetFromLUA))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"description": luaFuncTargetDescription,
		"dependent":   luaFuncTargetDependent,
	}))
	L.SetField(mod, pythonLibTargetLUAName, mt)
}

// Target defines the target interface
type Target interface {
	common.Object
	// Dependent on a dependency
	Dependent(dep *Dependency)
	// SetDescription sets description
	SetDescription(value string)
	// GetDescription gets description
	GetDescription() string
	// GetProto returns the protobuf object
	GetProto() *pbSpec.Target
}

// _Target defines the target
type _Target pbSpec.Target

// Dependent on a dependency
func (t *_Target) Dependent(dep *Dependency) {
	if dep != nil {
		t.Dependencies = append(t.Dependencies, (*pbSpec.Dependency)(dep))
	}
}

// SetDescription sets description
func (t *_Target) SetDescription(value string) {
	t.Description = value
}

// GetDescription gets description
func (t *_Target) GetDescription() string {
	return t.Description
}

// GetProto returns the protobuf object
func (t *_Target) GetProto() *pbSpec.Target {
	return (*pbSpec.Target)(t)
}

// GeneralTarget implements the general target
type GeneralTarget struct {
	_Target
}

// NewGeneralTargetFromLUA creates a new GTarget from LUA
func NewGeneralTargetFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	return new(GeneralTarget).GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (t *GeneralTarget) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = t
	L.SetMetatable(ud, L.GetTypeMetatable(targetLUATypeName))
	// Done
	return ud
}

// CommandTarget implements the command target
type CommandTarget struct {
	_Target
}

// NewCommandTargetFromLUA creates a new CommandTarget from LUA
func NewCommandTargetFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	command, err := params.GetString("command")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [command]: %v", err)
	}
	if command == "" {
		return nil, errors.New("Require command")
	}
	args, err := params.GetStringSlice("args")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [args]: %v", err)
	}
	workdir, err := params.GetString("workdir")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [workdir]: %v", err)
	}
	// Create a new Target
	target := &CommandTarget{
		_Target: _Target{
			Target: &pbSpec.Target_Command{
				Command: &pbSpec.CommandTarget{
					Command: command,
					Args:    args,
					Workdir: workdir,
				},
			},
		},
	}
	// Done
	return target.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (t *CommandTarget) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = t
	L.SetMetatable(ud, L.GetTypeMetatable(commandTargetLUATypeName))
	// Done
	return ud
}

// GoBinaryTarget implements the command target
type GoBinaryTarget struct {
	_Target
}

// NewGoBinaryTargetFromLUA creates a new GoBinaryTarget from LUA
func NewGoBinaryTargetFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	pkg, err := params.GetString("package")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [package]: %v", err)
	}
	if pkg == "" {
		return nil, errors.New("Require package")
	}
	output, err := params.GetString("output")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [output]: %v", err)
	}
	if output == "" {
		return nil, errors.New("Require output")
	}
	install, err := params.GetBool("install")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [install]: %v", err)
	}
	envs, err := params.GetStringSlice("envs")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [envs]: %v", err)
	}
	// Create a new target
	target := &GoBinaryTarget{
		_Target: _Target{
			Target: &pbSpec.Target_GoBinary{
				GoBinary: &pbSpec.GoBinaryTarget{
					Package: pkg,
					Output:  output,
					Install: install,
					Envs:    envs,
				},
			},
		},
	}
	// Done
	return target.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (t *GoBinaryTarget) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = t
	L.SetMetatable(ud, L.GetTypeMetatable(goBinaryTargetLUATypeName))
	// Done
	return ud
}

// PythonLibTarget implements the command target
type PythonLibTarget struct {
	_Target
}

// NewPythonLibTargetFromLUA creates a new PythonLibTarget from LUA
func NewPythonLibTargetFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	workdir, err := params.GetString("workdir")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [workdir]: %v", err)
	}
	setup, err := params.GetString("setup")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [setup]: %v", err)
	}
	// Create a new target
	target := &PythonLibTarget{
		_Target: _Target{
			Target: &pbSpec.Target_PythonLib{
				PythonLib: &pbSpec.PythonLibTarget{
					Workdir: workdir,
					Setup:   setup,
				},
			},
		},
	}
	// Done
	return target.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (t *PythonLibTarget) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = t
	L.SetMetatable(ud, L.GetTypeMetatable(pythonLibTargetLUATypeName))
	// Done
	return ud
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUATargetSelf get lua target self
func LUATargetSelf(L *lua.LState) Target {
	ud := L.CheckUserData(1)
	if target, ok := ud.Value.(Target); ok {
		return target
	}
	L.ArgError(1, "Target expected")
	return nil
}

// luaFuncTargetDescription defines target.description function in lua
func luaFuncTargetDescription(L *lua.LState) int {
	target := LUATargetSelf(L)
	if L.GetTop() == 1 {
		L.Push(lua.LString(target.GetDescription()))
		return 1
	} else if L.GetTop() == 2 {
		target.SetDescription(L.CheckString(2))
		return 0
	}
	L.ArgError(3, "Invalid argument")
	return 0
}

// luaFuncTargetDependent defines target.dependent function in lua
func luaFuncTargetDependent(L *lua.LState) int {
	target := LUATargetSelf(L)
	for i := 2; i <= L.GetTop(); i++ {
		ud := L.CheckUserData(i)
		dep, ok := ud.Value.(*Dependency)
		if !ok {
			L.ArgError(i, "Not a Dependency")
		}
		target.Dependent(dep)
	}
	// Done
	return 0
}

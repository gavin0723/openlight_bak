// Author: lipixun
// File Name: targets.go
// Description:

package build

import (
	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

// Exposed lua infos
const (
	LUATypeTarget            = "Build-Target"
	LUATypeCommandTarget     = "Build-Target-Command"
	LUATypeGoBinaryTarget    = "Build-Target-GoBinary"
	LUATypePythonLibTarget   = "Build-Target-PythonLib"
	LUATypeDockerImageTarget = "Build-Target-DockerImage"
)

// TargetLUAFuncs defines all lua functions for target
var TargetLUAFuncs = map[string]lua.LGFunction{
	"name":      LUAFuncTargetName,
	"dependent": LUAFuncTargetDependent,
	"options":   LUA.FuncObjectOptions,
}

// GoBinaryTargetLUAFuncs defines all lua functions for go binary target
var GoBinaryTargetLUAFuncs = map[string]lua.LGFunction{
	"package": LUAFuncGoBinaryTargetPackage,
}

// RegisterTargetType registers TargetType type in lua
func RegisterTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypeTarget)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), TargetLUAFuncs))
}

// RegisterCommandTargetType registers TargetType type in lua
func RegisterCommandTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypeTarget)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), TargetLUAFuncs))
}

// RegisterGoBinaryTargetType registers GoBinaryTarget type in lua
func RegisterGoBinaryTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypeGoBinaryTarget)
	var funcs = make(map[string]lua.LGFunction)
	for name, function := range TargetLUAFuncs {
		funcs[name] = function
	}
	for name, function := range GoBinaryTargetLUAFuncs {
		funcs[name] = function
	}
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), funcs))
}

// RegisterPythonLibTargetType registers PythonLibTarget type in lua
func RegisterPythonLibTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypePythonLibTarget)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), TargetLUAFuncs))
}

// Target represents the target of build package
type Target struct {
	LUA.Object
	Name string

	// Dependency
	Dependencies []Dependency

	// Target type
	Command     *CommandTarget
	GoBinary    *GoBinaryTarget
	PythonLib   *PythonLibTarget
	DockerImage *DockerImageTarget
}

// NewTarget creates a new Target
func NewTarget(luaTypeName, name string, options *lua.LTable) *Target {
	var target = Target{
		Name: name,
	}
	target.Object = LUA.NewObject(luaTypeName, options, &target)
	// Done
	return &target
}

// GetProto returns the protobuf object
func (target *Target) GetProto() (*pbSpec.Target, error) {
	var pbTarget pbSpec.Target
	pbTarget.Name = target.Name
	// Get dependencies
	for _, dep := range target.Dependencies {
		pbDeps, err := dep.GetProtos()
		if err != nil {
			return nil, err
		}
		pbTarget.Dependencies = append(pbTarget.Dependencies, pbDeps...)
	}
	// Check type
	if target.GoBinary != nil {
		// Go binary
		goBinary, err := target.GoBinary.GetProto(target.GetOptions())
		if err != nil {
			return nil, err
		}
		pbTarget.Target = &pbSpec.Target_GoBinary{
			GoBinary: goBinary,
		}
	} else if target.PythonLib != nil {
		// Python lib
		pythonLib, err := target.PythonLib.GetProto(target.GetOptions())
		if err != nil {
			return nil, err
		}
		pbTarget.Target = &pbSpec.Target_PythonLib{
			PythonLib: pythonLib,
		}
	} else if target.DockerImage != nil {
		// Docker image
		dockerImage, err := target.DockerImage.GetProto(target.GetOptions())
		if err != nil {
			return nil, err
		}
		pbTarget.Target = &pbSpec.Target_DockerImage{
			DockerImage: dockerImage,
		}
	}
	// Done
	return &pbTarget, nil
}

// CommandTarget represents the command target of build package
type CommandTarget struct {
	Args []string
}

// NewCommandTarget creates a new CommandTarget
func NewCommandTarget(name string, args []string, options *lua.LTable) *Target {
	target := NewTarget(LUATypeCommandTarget, name, options)
	target.Command = &CommandTarget{
		Args: args,
	}
	// Done
	return target
}

// GoBinaryTarget represents the go binary target of build package
type GoBinaryTarget struct {
	Package string
}

// NewGoBinaryTarget creates a new GoBinaryTarget
func NewGoBinaryTarget(name, pkg string, options *lua.LTable) *Target {
	target := NewTarget(LUATypeGoBinaryTarget, name, options)
	target.GoBinary = &GoBinaryTarget{
		Package: pkg,
	}
	// Done
	return target
}

// GetProto returns the protobuf object
func (target *GoBinaryTarget) GetProto(options *lua.LTable) (*pbSpec.GoBinaryTarget, error) {
	var err error
	var pbTarget pbSpec.GoBinaryTarget
	pbTarget.Package = target.Package
	if pbTarget.Output, err = LUA.TryGetStringFromTable(options, "output", ""); err != nil {
		return nil, err
	}
	return &pbTarget, nil
}

// PythonLibTarget represents the python library target of build package
type PythonLibTarget struct{}

// NewPythonLibTarget creates a new PythonLibTarget
func NewPythonLibTarget(name string, options *lua.LTable) *Target {
	target := NewTarget(LUATypePythonLibTarget, name, options)
	target.PythonLib = &PythonLibTarget{}
	// Done
	return target
}

// GetProto returns the protobuf object
func (target *PythonLibTarget) GetProto(options *lua.LTable) (*pbSpec.PythonLibTarget, error) {
	var err error
	var pbTarget pbSpec.PythonLibTarget
	if pbTarget.Setup, err = LUA.TryGetStringFromTable(options, "setup", ""); err != nil {
		return nil, err
	}
	if pbTarget.Workdir, err = LUA.TryGetStringFromTable(options, "workdir", ""); err != nil {
		return nil, err
	}
	return &pbTarget, nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUATargetSelf get lua target self
func LUATargetSelf(L *lua.LState) *Target {
	ud := L.CheckUserData(1)
	if ref, ok := ud.Value.(*Target); ok {
		return ref
	}
	L.ArgError(1, "Target expected")
	return nil
}

// LUAFuncTargetName defines target.name function in lua
func LUAFuncTargetName(L *lua.LState) int {
	target := LUATargetSelf(L)
	if target == nil {
		return 0
	}
	if L.GetTop() != 1 {
		L.ArgError(0, "Invalid arguments")
		return 0
	}
	// Return name
	L.Push(lua.LString(target.Name))
	return 1
}

// LUAFuncTargetDependent defines target.dependent function in lua
func LUAFuncTargetDependent(L *lua.LState) int {
	target := LUATargetSelf(L)
	if target == nil {
		return 0
	}
	var dependencies []Dependency
	for i := 2; i <= L.GetTop(); i++ {
		ud := L.CheckUserData(i)
		if ud == nil {
			return 0
		}
		dep, ok := ud.Value.(Dependency)
		if !ok {
			L.ArgError(i, "Not a dependency")
			return 0
		}
		dependencies = append(dependencies, dep)
	}
	target.Dependencies = append(target.Dependencies, dependencies...)
	// Done
	return 0
}

// LUAFuncGoBinaryTargetPackage defines GoBinary.package function in lua
func LUAFuncGoBinaryTargetPackage(L *lua.LState) int {
	target := LUATargetSelf(L)
	if target == nil {
		return 0
	}
	if target.GoBinary == nil {
		L.ArgError(0, "Not a GoBinary target")
		return 0
	}
	if L.GetTop() == 1 {
		// Get
		L.Push(lua.LString(target.DockerImage.Repository))
		return 1
	} else if L.GetTop() == 2 {
		// Set
		pkg := L.CheckString(2)
		if pkg == "" {
			L.ArgError(2, "Require value")
			return 0
		}
		target.GoBinary.Package = pkg
		return 0
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

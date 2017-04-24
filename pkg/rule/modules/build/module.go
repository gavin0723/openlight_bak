// Author: lipixun
// File Name: module.go
// Description:

package build

import (
	"fmt"

	"github.com/yuin/gopher-lua"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

const (
	// LUANameModule defines the name of BuildModule
	LUANameModule = "build"
)

// Ensure the interface is implements
var _ Module = (*_Module)(nil)

// RegisterTypes registers all types in this module
func RegisterTypes(L *lua.LState, mod *lua.LTable) {
	RegisterPackageType(L, mod)
	RegisterReferenceType(L, mod)
	RegisterPythonLocalFinderType(L, mod)
	RegisterTargetType(L, mod)
	RegisterCommandTargetType(L, mod)
	RegisterGoBinaryTargetType(L, mod)
	RegisterPythonLibTargetType(L, mod)
	RegisterDockerImageTargetType(L, mod)
	RegisterTargetDependencyType(L, mod)
	RegisterGoDependencyType(L, mod)
	RegisterPipDependencyType(L, mod)
	RegisterFileType(L, mod)
	RegisterArtifactType(L, mod)
}

// Module defines the module used by build progress
type Module interface {
	LUA.Module
	// Packages return all packages
	Packages() []*Package
}

// _Module implements the Module interface
type _Module struct {
	ctx      LUA.ModuleContext
	packages map[string]*Package
}

// NewModule returns new Module
func NewModule(ctx LUA.ModuleContext) Module {
	return &_Module{
		ctx:      ctx,
		packages: make(map[string]*Package),
	}
}

// InitLInitLUAModule initializes (preload) module into lua
func (m *_Module) InitLUAModule(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"getPackage":    m.LUAFuncGetPackage,
		"addPackage":    m.LUAFuncAddPackage,
		"deletePackage": m.LUAFuncDeletePackage,
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

// Packages return all packages
func (m *_Module) Packages() []*Package {
	var packages []*Package
	for _, pkg := range m.packages {
		packages = append(packages, pkg)
	}
	return packages
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUAFuncGetPackage defines build.getPackage in lua
func (m *_Module) LUAFuncGetPackage(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.ArgError(0, "Invalid arguments")
		return 0
	}
	pkg := m.packages[L.CheckString(1)]
	if pkg != nil {
		L.Push(pkg.GetLUAUserData(L))
		return 1
	}
	// Not found
	return 0
}

// LUAFuncAddPackage defines build.addPackage in lua
func (m *_Module) LUAFuncAddPackage(L *lua.LState) int {
	for i := 1; i <= L.GetTop(); i++ {
		ud := L.CheckUserData(i)
		if ud == nil {
			return 0
		}
		pkg, ok := ud.Value.(*Package)
		if !ok {
			L.ArgError(i, "Not a package")
			return 0
		}
		if _pkg := m.packages[pkg.Name]; _pkg != nil {
			L.ArgError(i, fmt.Sprintf("Duplicated package [%v]", _pkg.Name))
			return 0
		}
		m.packages[pkg.Name] = pkg
	}
	// Not found
	return 0
}

// LUAFuncDeletePackage defines build.deletePackage in lua
func (m *_Module) LUAFuncDeletePackage(L *lua.LState) int {
	for i := 1; i < L.GetTop(); i++ {
		delete(m.packages, L.CheckString(i))
	}
	// Done
	return 0
}

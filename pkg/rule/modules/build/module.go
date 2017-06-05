// Author: lipixun
// File Name: module.go
// Description:

package build

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
	ModuleName = "build"
)

const (
	moduleLUAName = "build"
)

// Ensure the interface is implements
var _ engine.Module = (*_Module)(nil)
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

// Module defines the module used by build progress
type Module interface {
	engine.Module
	// Spec returns the build file spec
	Spec() *pbSpec.BuildFile
}

// _Module implements the Module interface
type _Module struct {
	ctx *engine.Context
	pkg *Package
}

// NewModule returns new Module
func NewModule(ctx *engine.Context) Module {
	return &_Module{ctx: ctx}
}

// LUAName returns the module name in lua
func (m *_Module) LUAName() string {
	return moduleLUAName
}

// InitLInitLUAModule initializes module into lua
func (m *_Module) InitLUAModule(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"package": m.luaPackage,
	})
	// Register types
	registerPackageType(L, mod)
	registerReferenceType(L, mod)
	registerPythonFinderType(L, mod)
	registerGoFinderType(L, mod)
	registerTargetType(L, mod)
	registerCommandTargetType(L, mod)
	registerGoBinaryTargetType(L, mod)
	registerPythonLibTargetType(L, mod)
	registerDockerImageTargetType(L, mod)
	registerTargetDependencyType(L, mod)
	registerGoDependencyType(L, mod)
	registerPipDependencyType(L, mod)
	registerFileType(L, mod)
	registerArtifactType(L, mod)
	// Return
	L.Push(mod)
	return 1
}

// Spec returns the build file spec
func (m *_Module) Spec() *pbSpec.BuildFile {
	var buildfile pbSpec.BuildFile
	if m.pkg != nil {
		buildfile.Package = m.pkg.GetProto()
	}
	return &buildfile
}

// SetRuleFiles sets the rule files by this module
func (m *_Module) SetRuleFiles(ruleFile *pbSpec.RuleFiles) error {
	ruleFile.Build = m.Spec()
	return nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// luaPackage returns package
func (m *_Module) luaPackage(L *lua.LState) int {
	if L.GetTop() == 0 {
		L.ArgError(0, "Require argument")
	} else if L.GetTop() > 1 {
		L.ArgError(2, "Too many arguments")
	}
	// Check package
	if m.pkg != nil {
		L.ArgError(0, "Package has already declared")
	}
	// Get parameters
	params, err := common.NewParametersFromLUATable(L.CheckTable(1))
	if err != nil {
		L.ArgError(1, "Invalid parameter")
	}
	// Create package
	name, err := params.GetString("name")
	if err != nil {
		L.ArgError(1, fmt.Sprintf("Invalid parameter [name]: %v", err))
	}
	if name == "" {
		L.ArgError(1, "Require parameter [name]")
	}
	// Create metas
	meta := make(map[string]string)
	for key, value := range params {
		if key != "name" {
			if value.Type() != lua.LTString {
				L.ArgError(1, fmt.Sprintf("Invalid parameter [%v], must be a string", key))
			}
			meta[key] = string(value.(lua.LString))
		}
	}
	// Create package
	pkg := NewPackage(name, meta)
	m.pkg = pkg
	// Return
	L.Push(pkg.GetLUAUserData(L))
	return 1
}

// Author: lipixun
// File Name: dependency.go
// Description:

package build

import (
	"github.com/yuin/gopher-lua"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// Exposed lua infos
const (
	LUANameTargetDependency = "TargetDependency"
	LUANameGoDependency     = "GoDependency"
	LUANamePipDependency    = "PipDependency"

	LUATypeTargetDependency = "Build-Dependency-Target"
	LUATypeGoDependency     = "Build-Dependency-Go"
	LUATypePipDependency    = "Build-Dependency-Pip"
)

// TargetDependencyLUAFuncs defines all lua functions for TargetDependency
var TargetDependencyLUAFuncs = map[string]lua.LGFunction{
	"options": LUA.FuncObjectOptions,
}

// GoDependencyLUAFuncs defines all lua functions for GoDependency
var GoDependencyLUAFuncs = map[string]lua.LGFunction{
	"options": LUA.FuncObjectOptions,
}

// PipDependencyLUAFuncs defines all lua functions for PipDependency
var PipDependencyLUAFuncs = map[string]lua.LGFunction{
	"options": LUA.FuncObjectOptions,
}

// RegisterTargetDependencyType registers target dependency type in lua
func RegisterTargetDependencyType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypeTargetDependency)
	L.SetField(mt, "new", L.NewFunction(LUAFuncTargetDependencyNew))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), TargetDependencyLUAFuncs))
	L.SetField(mod, LUANameTargetDependency, mt)
}

// RegisterGoDependencyType registers target dependency type in lua
func RegisterGoDependencyType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypeGoDependency)
	L.SetField(mt, "new", L.NewFunction(LUAFuncGoDependencyNew))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), GoDependencyLUAFuncs))
	L.SetField(mod, LUANameGoDependency, mt)
}

// RegisterPipDependencyType registers target dependency type in lua
func RegisterPipDependencyType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypePipDependency)
	L.SetField(mt, "new", L.NewFunction(LUAFuncPipDependencyNew))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), PipDependencyLUAFuncs))
	L.SetField(mod, LUANamePipDependency, mt)
}

// Dependency defines the dependency interface
type Dependency interface {
	LUA.Object
	// GetProto returns the protobuf object
	GetProtos() ([]*pbSpec.Dependency, error)
}

func getDependencyOptionsProto(options *lua.LTable) (*pbSpec.DependencyOptions, error) {
	var pbOptions pbSpec.DependencyOptions
	if options != nil {
		build, err := LUA.TryGetBoolFromTable(options, "build", false)
		if err != nil {
			return nil, err
		}
		if build {
			pbOptions.Build = true
		}
	}
	// Done
	return &pbOptions, nil
}

// TargetDependency represents the target dependency
type TargetDependency struct {
	LUA.Object
	Package string
	Target  string
}

// NewTargetDependency creates a new TargetDependency
func NewTargetDependency(pkg, target string, options *lua.LTable) *TargetDependency {
	var dep = TargetDependency{
		Package: pkg,
		Target:  target,
	}
	dep.Object = LUA.NewObject(LUATypeTargetDependency, options, Dependency(&dep))
	// Done
	return &dep
}

// GetProtos returns the protobuf object
func (dep *TargetDependency) GetProtos() ([]*pbSpec.Dependency, error) {
	var pbDep pbSpec.Dependency
	pbDep.Dependency = &pbSpec.Dependency_Target{
		Target: &pbSpec.TargetDependency{
			Package: dep.Package,
			Target:  dep.Target,
		},
	}
	var err error
	pbDep.Options, err = getDependencyOptionsProto(dep.GetOptions())
	if err != nil {
		return nil, err
	}
	// Done
	return []*pbSpec.Dependency{&pbDep}, nil
}

// GoDependency represents the go dependency
type GoDependency struct {
	LUA.Object
	Packages []string
}

// NewGoDependency creates a new Dependency
func NewGoDependency(packages []string, options *lua.LTable) *GoDependency {
	var dep = GoDependency{
		Packages: packages,
	}
	dep.Object = LUA.NewObject(LUATypeGoDependency, options, Dependency(&dep))
	// Done
	return &dep
}

// GetProtos returns the protobuf object
func (dep *GoDependency) GetProtos() ([]*pbSpec.Dependency, error) {
	var pbDeps []*pbSpec.Dependency
	// Get options
	pbOptions, err := getDependencyOptionsProto(dep.GetOptions())
	if err != nil {
		return nil, err
	}
	// Get dependencies
	var pkgNames = make(map[string]bool)
	for _, pkg := range dep.Packages {
		if pkg != "" && !pkgNames[pkg] {
			pbDeps = append(pbDeps, &pbSpec.Dependency{
				Dependency: &pbSpec.Dependency_Go{
					Go: &pbSpec.GoDependency{
						Package: pkg,
					},
				},
				Options: pbOptions,
			})
			pkgNames[pkg] = true
		}
	}
	// Done
	return pbDeps, nil
}

// PipDependency represents the pip dependency
type PipDependency struct {
	LUA.Object
	Modules []string
}

// NewPipDependency creates a new Dependency
func NewPipDependency(modules []string, options *lua.LTable) *PipDependency {
	var dep = PipDependency{
		Modules: modules,
	}
	dep.Object = LUA.NewObject(LUATypePipDependency, options, Dependency(&dep))
	// Done
	return &dep
}

// GetProtos returns the protobuf object
func (dep *PipDependency) GetProtos() ([]*pbSpec.Dependency, error) {
	var pbDeps []*pbSpec.Dependency
	// Get options
	pbOptions, err := getDependencyOptionsProto(dep.GetOptions())
	if err != nil {
		return nil, err
	}
	// Get dependencies
	var moduleNames = make(map[string]bool)
	for _, module := range dep.Modules {
		if module != "" && !moduleNames[module] {
			pbDeps = append(pbDeps, &pbSpec.Dependency{
				Dependency: &pbSpec.Dependency_Pip{
					Pip: &pbSpec.PipDependency{
						Module: module,
					},
				},
				Options: pbOptions,
			})
			moduleNames[module] = true
		}
	}
	// Done
	return pbDeps, nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUAFuncTargetDependencyNew defines TargetDependency.new function in lua
func LUAFuncTargetDependencyNew(L *lua.LState) int {
	// Get repository or target
	pkgOrTarget := L.CheckString(1)
	if pkgOrTarget == "" {
		L.ArgError(1, "Require name")
		return 0
	}
	if L.GetTop() == 1 {
		// Target without options
		target := NewTargetDependency("", pkgOrTarget, nil)
		// Return
		L.Push(target.GetLUAUserData(L))
		return 1
	}
	// Check the second argument
	value2 := L.Get(2)
	switch value2.Type() {
	case lua.LTString:
		// Target
		target := NewTargetDependency(pkgOrTarget, string(value2.(lua.LString)), L.ToTable(3))
		L.Push(target.GetLUAUserData(L))
		return 1
	case lua.LTTable:
		// Options
		target := NewTargetDependency("", pkgOrTarget, L.ToTable(2))
		L.Push(target.GetLUAUserData(L))
		return 1
	default:
		L.TypeError(2, value2.Type())
		return 0
	}
}

// LUAFuncGoDependencyNew defines GoDependency.new function in lua
func LUAFuncGoDependencyNew(L *lua.LState) int {
	// Get packages
	if L.GetTop() == 0 {
		L.ArgError(0, "Require package")
		return 0
	}
	var packages []string
	var packageMap = make(map[string]bool)
	for i := 1; i <= L.GetTop()-1; i++ {
		pkg := L.CheckString(i)
		if pkg == "" {
			L.ArgError(i, "Package cannot be empty")
			return 0
		}
		if packageMap[pkg] {
			L.ArgError(i, "Duplicated package")
			return 0
		}
		packages = append(packages, pkg)
		packageMap[pkg] = true
	}
	var options *lua.LTable
	value := L.Get(L.GetTop())
	if value.Type() == lua.LTString {
		pkg := value.(lua.LString)
		if pkg == "" {
			L.ArgError(L.GetTop(), "Package cannot be empty")
		}
		packages = append(packages, string(pkg))
	} else if value.Type() == lua.LTTable {
		options = value.(*lua.LTable)
	} else {
		L.TypeError(L.GetTop(), value.Type())
		return 0
	}
	// Create go dependency
	dep := NewGoDependency(packages, options)
	// Return
	L.Push(dep.GetLUAUserData(L))
	return 1
}

// LUAFuncPipDependencyNew defines PipDependency.new function in lua
func LUAFuncPipDependencyNew(L *lua.LState) int {
	// Get modules
	if L.GetTop() == 0 {
		L.ArgError(0, "Require module")
		return 0
	}
	var modules []string
	var moduleMap = make(map[string]bool)
	for i := 1; i <= L.GetTop()-1; i++ {
		module := L.CheckString(i)
		if module == "" {
			L.ArgError(i, "Module cannot be empty")
			return 0
		}
		if moduleMap[module] {
			L.ArgError(i, "Duplicated module")
			return 0
		}
		modules = append(modules, module)
		moduleMap[module] = true
	}
	var options *lua.LTable
	value := L.Get(L.GetTop())
	if value.Type() == lua.LTString {
		module := value.(lua.LString)
		if module == "" {
			L.ArgError(L.GetTop(), "Module cannot be empty")
		}
		modules = append(modules, string(module))
	} else if value.Type() == lua.LTTable {
		options = value.(*lua.LTable)
	} else {
		L.TypeError(L.GetTop(), value.Type())
		return 0
	}
	// Create pip dependency
	dep := NewPipDependency(modules, options)
	// Return
	L.Push(dep.GetLUAUserData(L))
	return 1
}

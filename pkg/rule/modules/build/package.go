// Author: lipixun
// File Name: package.go
// Description:
// 	Lua package object has the following fields and functions:
//	Fields:
//		name
//		remote
//		options
// Functions:
//		reference
//

package build

import (
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

// Exposed lua infos
const (
	LUANamePackage = "Package"
	LUATypePackage = "Build-Package"
)

// PackageLUAFuncs defines all lua functions for package
var PackageLUAFuncs = map[string]lua.LGFunction{
	"name":        LUAFuncPackageName,
	"remote":      LUAFuncPackageRemote,
	"options":     LUA.FuncObjectOptions,
	"reference":   LUAFuncPackageReference,
	"target":      LUAFuncPackageTarget,
	"command":     LUAFuncPackageCommand,
	"goBinary":    LUAFuncPackageGoBinary,
	"pythonLib":   LUAFuncPackagePythonLib,
	"dockerImage": LUAFuncPackageDockerImage,
}

// RegisterPackageType registers package type in lua
func RegisterPackageType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(LUATypePackage)
	L.SetField(mt, "new", L.NewFunction(LUAPackageNew))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), PackageLUAFuncs))
	// Add type to module
	L.SetField(mod, LUANamePackage, mt)
}

// Package represents the build package
type Package struct {
	LUA.Object
	Name       string
	Remote     string
	References map[string]*Reference
	Targets    map[string]*Target
}

// NewPackage creates a new Package
func NewPackage(name, remote string, references map[string]*Reference, targets map[string]*Target, options *lua.LTable) *Package {
	if references == nil {
		references = make(map[string]*Reference)
	}
	if targets == nil {
		targets = make(map[string]*Target)
	}
	var pkg = Package{
		Name:       name,
		Remote:     remote,
		References: references,
		Targets:    targets,
	}
	pkg.Object = LUA.NewObject(LUATypePackage, options, &pkg)
	// Done
	return &pkg
}

// GetProto returns the protobuf object
func (pkg *Package) GetProto() (*pbSpec.Package, error) {
	var pbPackage pbSpec.Package
	pbPackage.Name = pkg.Name
	pbPackage.Remote = pkg.Remote
	// Get references
	if len(pkg.References) > 0 {
		pbPackage.References = make(map[string]*pbSpec.Reference)
		for name, ref := range pkg.References {
			pbReference, err := ref.GetProto()
			if err != nil {
				return nil, err
			}
			pbPackage.References[name] = pbReference
		}
	}
	// Get targets
	if len(pkg.Targets) > 0 {
		pbPackage.Targets = make(map[string]*pbSpec.Target)
		for name, target := range pkg.Targets {
			pbTarget, err := target.GetProto()
			if err != nil {
				return nil, err
			}
			pbPackage.Targets[name] = pbTarget
		}
	}
	// Get options
	var pbOptions pbSpec.PackageOptions
	if options := pkg.GetOptions(); options != nil {
		var err error
		if pbOptions.DefaultTargets, err = LUA.GetStringSliceFromTable(options, "defaultTargets"); err != nil {
			return nil, fmt.Errorf("Failed to get options [defaultTargets], error: %s", err)
		}
	}
	pbPackage.Options = &pbOptions
	// Done
	return &pbPackage, nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUAPackageNew defines package.new function in lua
func LUAPackageNew(L *lua.LState) int {
	name := L.CheckString(1)
	if name == "" {
		L.ArgError(1, "Require name")
		return 0
	}
	pkg := NewPackage(name, L.CheckString(2), nil, nil, L.CheckTable(3))
	// Return
	L.Push(pkg.GetLUAUserData(L))
	return 1
}

// LUAPackageSelf get lua package self
func LUAPackageSelf(L *lua.LState) *Package {
	ud := L.CheckUserData(1)
	if pkg, ok := ud.Value.(*Package); ok {
		return pkg
	}
	L.ArgError(1, "Package expected")
	return nil
}

// LUAFuncPackageName defines package.name function in lua
func LUAFuncPackageName(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	if L.GetTop() != 1 {
		L.ArgError(0, "Invalid arguments")
		return 0
	}
	// Return name
	L.Push(lua.LString(pkg.Name))
	return 1
}

// LUAFuncPackageRemote defines package.remote function in lua
func LUAFuncPackageRemote(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	if L.GetTop() == 1 {
		// Get
		L.Push(lua.LString(pkg.Remote))
		return 1
	} else if L.GetTop() == 2 {
		// Set
		pkg.Remote = L.CheckString(2)
		return 0
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

// LUAFuncPackageReference defines package.reference in lua
func LUAFuncPackageReference(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	if L.GetTop() == 2 {
		// Find references
		ref := pkg.References[L.CheckString(2)]
		if ref == nil {
			return 0
		}
		L.Push(ref.GetLUAUserData(L))
		return 1
	}
	// Create new references
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	ref := pkg.References[name]
	if ref != nil {
		L.ArgError(2, fmt.Sprintf("Duplicated reference name [%v]", name))
		return 0
	}
	ref = NewReference(name, L.CheckString(3), nil, L.ToTable(4))
	pkg.References[name] = ref
	// Return
	L.Push(ref.GetLUAUserData(L))
	return 1
}

// LUAFuncPackageTarget defines package.target in lua
func LUAFuncPackageTarget(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	target := pkg.Targets[name]
	if target != nil {
		L.Push(target.GetLUAUserData(L))
		return 1
	}
	// Create a new target
	target = NewTarget(LUATypeTarget, name, L.ToTable(3))
	pkg.Targets[name] = target
	// Return
	L.Push(target.GetLUAUserData(L))
	return 1
}

// LUAFuncPackageCommand defines package.command in lua
func LUAFuncPackageCommand(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	target := pkg.Targets[name]
	if target != nil {
		L.ArgError(2, fmt.Sprintf("Duplicated target name [%v]", name))
		return 0
	}
	// Get args
	var args []string
	for i := 2; i <= L.GetTop()-1; i++ {
		args = append(args, L.CheckString(i))
	}
	var options *lua.LTable
	value := L.Get(L.GetTop())
	if value.Type() == lua.LTString {
		args = append(args, string(value.(lua.LString)))
	} else if value.Type() == lua.LTTable {
		options = value.(*lua.LTable)
	} else {
		L.TypeError(L.GetTop(), value.Type())
		return 0
	}
	// Create command target
	target = NewCommandTarget(name, args, options)
	pkg.Targets[name] = target
	// Return
	L.Push(target.GetLUAUserData(L))
	return 1
}

// LUAFuncPackageGoBinary defines package.goBinary in lua
func LUAFuncPackageGoBinary(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	// Get name
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	target := pkg.Targets[name]
	if target != nil {
		L.ArgError(2, fmt.Sprintf("Duplicated target name [%v]", name))
		return 0
	}
	goPackage := L.CheckString(3)
	if goPackage == "" {
		L.ArgError(3, "Require package")
		return 0
	}
	// Create target
	target = NewGoBinaryTarget(name, goPackage, L.ToTable(4))
	pkg.Targets[name] = target
	// Return
	L.Push(target.GetLUAUserData(L))
	return 1
}

// LUAFuncPackagePythonLib defines package.pythonLib in lua
func LUAFuncPackagePythonLib(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	// Get name
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	target := pkg.Targets[name]
	if target != nil {
		L.ArgError(2, fmt.Sprintf("Duplicated target name [%v]", name))
		return 0
	}
	// Create target
	target = NewPythonLibTarget(name, L.ToTable(3))
	pkg.Targets[name] = target
	// Return
	L.Push(target.GetLUAUserData(L))
	return 1
}

// LUAFuncPackageDockerImage defines package.dockerImage in lua
func LUAFuncPackageDockerImage(L *lua.LState) int {
	pkg := LUAPackageSelf(L)
	if pkg == nil {
		return 0
	}
	// Get name
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	target := pkg.Targets[name]
	if target != nil {
		L.ArgError(2, fmt.Sprintf("Duplicated target name [%v]", name))
		return 0
	}
	repository := L.CheckString(3)
	if repository == "" {
		L.ArgError(3, "Require repository")
		return 0
	}
	imageName := L.CheckString(4)
	if imageName == "" {
		L.ArgError(4, "Require image name")
		return 0
	}
	// Create target
	target = NewDockerImageTarget(name, repository, imageName, L.ToTable(5))
	pkg.Targets[name] = target
	// Return
	L.Push(target.GetLUAUserData(L))
	return 1
}

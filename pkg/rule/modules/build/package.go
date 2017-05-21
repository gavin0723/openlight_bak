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
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/rule/common"
)

// Ensure the interface is implemented
var _ common.Object = (*Package)(nil)

const (
	packageLUAName     = "Package"
	packageLUATypeName = "build-package"
)

// registerPackageType registers package type in lua
func registerPackageType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(packageLUATypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"references": luaFuncPackageReferences,
		"targets":    luaFuncPackageTargets,
		"options":    luaFuncPackageOptions,
	}))
}

// Package represents the build package
type Package struct {
	name       string
	meta       map[string]string
	references *common.NamedCollection
	targets    *common.NamedCollection
	options    *common.Options
}

// NewPackage creates a new Package
func NewPackage(name string, meta map[string]string) *Package {
	return &Package{name: name, meta: meta}
}

// GetLUAUserData returns the lua user data
func (pkg *Package) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = pkg
	L.SetMetatable(ud, L.GetTypeMetatable(packageLUATypeName))
	// Done
	return ud
}

// GetProto returns the protobuf object
func (pkg *Package) GetProto() *pbSpec.Package {
	var pbPackage pbSpec.Package
	pbPackage.Name = pkg.name
	pbPackage.Meta = pkg.meta
	// References
	if pkg.references != nil {
		pbPackage.References = make(map[string]*pbSpec.Reference)
		for _, item := range pkg.references.Items() {
			pbPackage.References[item.Name] = item.Value.(*Reference).GetProto()
		}
	}
	// Targets
	if pkg.targets != nil {
		pbPackage.Targets = make(map[string]*pbSpec.Target)
		for _, item := range pkg.targets.Items() {
			pbPackage.Targets[item.Name] = item.Value.(Target).GetProto()
		}
	}
	// Options
	if pkg.options != nil {
		pbPackage.Options = new(pbSpec.PackageOptions)
		value := pkg.options.Get("defaultTargets")
		if value != nil {
			pbPackage.Options.DefaultTargets = value.Value().([]string)
		}
	}
	// Done
	return &pbPackage
}

func convertReferenceValue(value lua.LValue) (common.Object, error) {
	if value == nil {
		return nil, errors.New("Invalid nil value")
	}
	if value.Type() != lua.LTUserData {
		return nil, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTUserData, value.Type())
	}
	ref, ok := value.(*lua.LUserData).Value.(*Reference)
	if !ok {
		return nil, errors.New("Expect Reference object")
	}
	return ref, nil
}

func convertTargetValue(value lua.LValue) (common.Object, error) {
	if value == nil {
		return nil, errors.New("Invalid nil value")
	}
	if value.Type() != lua.LTUserData {
		return nil, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTUserData, value.Type())
	}
	target, ok := value.(*lua.LUserData).Value.(Target)
	if !ok {
		return nil, errors.New("Expect Target object")
	}
	return target, nil
}

func convertOptionValue(key string, value lua.LValue) (interface{}, error) {
	switch key {
	case "defaultTargets": // A string slice
		return common.ConvertValueToStringSlice(value)
	}
	// Bad
	return nil, fmt.Errorf("Unknown option key [%v]", key)
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

func luaPackageSelf(L *lua.LState) *Package {
	ud := L.CheckUserData(1)
	if pkg, ok := ud.Value.(*Package); ok {
		return pkg
	}
	L.ArgError(1, "Package expected")
	return nil
}

func luaFuncPackageReferences(L *lua.LState) int {
	pkg := luaPackageSelf(L)
	if pkg.references == nil {
		pkg.references = common.NewNamedCollection(convertReferenceValue)
	}
	L.Push(pkg.references.GetLUAUserData(L))
	return 1
}

func luaFuncPackageTargets(L *lua.LState) int {
	pkg := luaPackageSelf(L)
	if pkg.targets == nil {
		pkg.targets = common.NewNamedCollection(convertTargetValue)
	}
	L.Push(pkg.targets.GetLUAUserData(L))
	return 1
}

func luaFuncPackageOptions(L *lua.LState) int {
	pkg := luaPackageSelf(L)
	if pkg.options == nil {
		pkg.options = common.NewOptions(convertOptionValue)
	}
	L.Push(pkg.options.GetLUAUserData(L))
	return 1
}

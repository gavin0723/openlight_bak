// Author: lipixun
// Created Time : 日  3/12 15:41:16 2017
//
// File Name: lua.go
// Description:
//
//	The functions which are exposed to lua build spec file

package spec

import (
	"fmt"

	"github.com/yuin/gopher-lua"
)

// LUAFunctionProtoType The lua function proto type
type LUAFunctionProtoType interface {
	// Return if this function is safe
	Safe() bool
	// Get the exposed lua function to a specific parse context
	LUAFunction(ctx *ParseContext) lua.LGFunction
}

// LGContextFunction The lua function with parse context
type LGContextFunction func(ctx *ParseContext, l *lua.LState) int

// LUAFunctions The functions exposed to lua
var LUAFunctions = map[string]LUAFunctionProtoType{
	"package":           NewGeneralLUAFunctionProtoType(true, luaPackage),
	"options":           NewGeneralLUAFunctionProtoType(true, luaOptions),
	"reference":         NewGeneralLUAFunctionProtoType(true, luaReference),
	"localPythonFinder": NewGeneralLUAFunctionProtoType(true, luaLocalPythonFinder),
	"targetOptions":     NewGeneralLUAFunctionProtoType(true, luaTargetOptions),
	"goBinary":          NewGeneralLUAFunctionProtoType(true, luaGoBinary),
	"pythonLib":         NewGeneralLUAFunctionProtoType(true, luaPythonLib),
	"targetDependency":  NewGeneralLUAFunctionProtoType(true, luaTargetDependency),
	"goDependency":      NewGeneralLUAFunctionProtoType(true, luaGoDependency),
	"pipDependency":     NewGeneralLUAFunctionProtoType(true, luaPipDependency),
}

// GeneralLUAFunctionProtoType The general lua function proto type
type GeneralLUAFunctionProtoType struct {
	isSafe   bool
	function LGContextFunction
}

// NewGeneralLUAFunctionProtoType Create a new GeneralLUAFunctionProtoType
func NewGeneralLUAFunctionProtoType(isSafe bool, function LGContextFunction) *GeneralLUAFunctionProtoType {
	return &GeneralLUAFunctionProtoType{isSafe: isSafe, function: function}
}

// Safe Is safe
func (g *GeneralLUAFunctionProtoType) Safe() bool {
	return g.isSafe
}

// LUAFunction Get the lua function
func (g *GeneralLUAFunctionProtoType) LUAFunction(ctx *ParseContext) lua.LGFunction {
	if g.function == nil {
		return nil
	}
	return func(l *lua.LState) int {
		return g.function(ctx, l)
	}
}

// NewOptionsFromTable Create new options from table
func NewOptionsFromTable(ctx *ParseContext, l *lua.LState, table *lua.LTable) Options {
	if table == nil {
		return nil
	}
	options := NewOptions()
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if key.Type() != lua.LTString {
			// Bad key
			panic(fmt.Sprintf("Option key must be a string. Got [%s] type [%v].", key.String(), key.Type()))
		}
		keyStr := key.String()
		switch value.(type) {
		case lua.LBool:
			options.Set(keyStr, bool(value.(interface{}).(lua.LBool)))
		case lua.LString:
			options.Set(keyStr, value.String())
		case lua.LNumber:
			options.Set(keyStr, float64(value.(interface{}).(lua.LNumber)))
		case *lua.LTable:
			if v := NewOptionsOrSliceFromTable(ctx, l, value.(*lua.LTable)); v != nil {
				options.Set(keyStr, v)
			}
		default:
			panic(fmt.Sprintf("Unsupported value type [%v] in options", value.Type()))
		}
	})
	// Done
	return options
}

// NewOptionsOrSliceFromTable Create new options or slice from table
func NewOptionsOrSliceFromTable(ctx *ParseContext, l *lua.LState, table *lua.LTable) interface{} {
	if table == nil {
		return nil
	}
	if table.RawGetInt(1) != lua.LNil {
		// A slice, only support string slice
		slice := make([]string, table.Len(), table.Len())
		table.ForEach(func(key lua.LValue, value lua.LValue) {
			if key.Type() != lua.LTNumber {
				// Bad key
				panic(fmt.Sprintf("Option key must be a string or bad array index: [%s] type [%v].", key.String(), key.Type()))
			}
			index := int(key.(interface{}).(lua.LNumber)) - 1
			if index >= table.Len() {
				// Bad key
				panic(fmt.Sprintf("Option key must be a string or bad array index: [%s] type [%v].", key.String(), key.Type()))
			}
			if value.Type() != lua.LTString {
				// Bad value
				panic(fmt.Sprintf("Slice value must be a string: [%s] type [%v]", value.String(), value.Type()))
			}
			// Set
			slice[index] = value.String()
		})
		// Done
		return slice
	}
	// An options
	return NewOptionsFromTable(ctx, l, table)
}

// LUA function: package
// Parameter:
//	name 				The package name. A string. Required.
//	options 			The package options. A table. Optional.
func luaPackage(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	if buildSpec.Package != nil {
		panic("Multiple package declaration found.")
	}
	// Create package
	buildSpec.Package = new(PackageSpec)
	// Get parameters
	name := l.ToString(1)
	if name == "" {
		panic("Require package name")
	}
	buildSpec.Package.Name = name
	buildSpec.Package.Options = NewOptionsFromTable(ctx, l, l.ToTable(2))
	// Done
	return 0
}

// LUA function: options
// Parameter:
//	options 			The package options. A table. Optional.
// NOTE:
//	Must be declared after package
func luaOptions(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	if buildSpec.Package == nil {
		panic("Package must be declared before options.")
	}
	// Parse options
	options := NewOptionsFromTable(ctx, l, l.ToTable(1))
	if options != nil {
		if buildSpec.Package.Options == nil {
			buildSpec.Package.Options = NewOptions()
		}
		buildSpec.Package.Options.Merge(options)
	}
	// Done
	return 0
}

// LUA function: reference
// Parameter:
//	name 				The package name. A string. Required.
//	remoteURI 			The remote uri. A string. Optional.
func luaReference(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Create reference
	var referenceSpec PackageReferenceSpec
	name := l.ToString(1)
	if name == "" {
		panic("Require reference package name.")
	}
	if spec := buildSpec.GetReference(name); spec != nil {
		panic(fmt.Sprintf("Duplicated reference to [%s]", name))
	}
	referenceSpec.Name = name
	referenceSpec.Remote = l.ToString(2)
	// Add
	buildSpec.References = append(buildSpec.References, &referenceSpec)
	// Done
	return 0
}

// LUA function: localPythonFinder
// Parameter:
//	name 				The package name. A string. Required.
//	module 				The python module name. A string. Required.
// 	options 			The options. A table. Optional.
func luaLocalPythonFinder(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Get name
	name := l.ToString(1)
	if name == "" {
		panic("Require name of python local finder.")
	}
	// Find the reference
	referenceSpec := buildSpec.GetReference(name)
	if referenceSpec == nil {
		panic("Reference must be declared before finder.")
	}
	// Get module
	module := l.ToString(2)
	if module == "" {
		panic("Require module of python local finder.")
	}
	// Create finder
	var finder PackageLocalPythonFinderSpec
	finder.Module = module
	finder.Options = NewOptionsFromTable(ctx, l, l.ToTable(3))
	// Add
	referenceSpec.Finders = append(referenceSpec.Finders, &PackageFinderSpec{
		Type:        PackageFinderTypeLocalPython,
		LocalPython: &finder,
	})
	// Done
	return 0
}

// LUA function: goBinary
// Parameter:
//	name 				The target name. A string. Required.
//	package 			The go package to build. A string. Required.
// 	options 			The options. A table. Optional.
func luaGoBinary(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Get name
	name := l.ToString(1)
	if name == "" {
		panic("Require target name.")
	}
	if spec := buildSpec.GetTarget(name); spec != nil {
		panic(fmt.Sprintf("Duplicated target [%s]", name))
	}
	pkg := l.ToString(2)
	if pkg == "" {
		panic("Require package name of go binary target.")
	}
	// Create target
	var target TargetSpec
	target.Name = name
	target.Type = TargetTypeGoBinary
	target.GoBinary = &GoBinaryTargetSpec{Package: pkg}
	target.Options = NewOptionsFromTable(ctx, l, l.ToTable(3))
	// Add
	buildSpec.Targets = append(buildSpec.Targets, &target)
	// Done
	return 0
}

// LUA function: pythonLib
// Parameter:
//	name 				The target name. A string. Required.
//	package 			The go package to build. A string. Required.
// 	options 			The options. A table. Optional.
func luaPythonLib(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Get name
	name := l.ToString(1)
	if name == "" {
		panic("Require target name.")
	}
	if spec := buildSpec.GetTarget(name); spec != nil {
		panic(fmt.Sprintf("Duplicated target [%s]", name))
	}
	// Create target
	var target TargetSpec
	target.Name = name
	target.Type = TargetTypePythonLib
	target.PythonLib = &PythonLibTargetSpec{}
	target.Options = NewOptionsFromTable(ctx, l, l.ToTable(3))
	// Add
	buildSpec.Targets = append(buildSpec.Targets, &target)
	// Done
	return 0
}

// LUA function: targetOptions
// Parameter:
//	name 				The target name. A string. Required.
// 	options 			The options. A table. Optional.
func luaTargetOptions(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Get name
	name := l.ToString(1)
	if name == "" {
		panic("Require target name.")
	}
	targetSpec := buildSpec.GetTarget(name)
	if targetSpec == nil {
		panic("Target must be declared before target options")
	}
	// Get options
	options := NewOptionsFromTable(ctx, l, l.ToTable(1))
	if options != nil {
		if targetSpec.Options == nil {
			targetSpec.Options = NewOptions()
		}
		targetSpec.Options.Merge(options)
	}
	// Done
	return 0
}

// LUA function: targetDependency
// Parameter:
//	name 				The target name. A string. Required.
// 	reference 			The reference uri. A string. Required.
// 	options 			The options. A table. Optional.
func luaTargetDependency(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Get name
	name := l.ToString(1)
	if name == "" {
		panic("Require target name.")
	}
	targetSpec := buildSpec.GetTarget(name)
	if targetSpec == nil {
		panic("Target must be declared before target options")
	}
	// Get reference
	referenceStr := l.ToString(2)
	if referenceStr == "" {
		panic("Require target reference.")
	}
	referenceSpec, err := ParseReferenceFromString(referenceStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid target reference: %s", err))
	}
	// Create target dependency
	var targetDependency DependencySpec
	targetDependency.Reference = referenceSpec
	targetDependency.Options = NewOptionsFromTable(ctx, l, l.ToTable(3))
	// Add
	targetSpec.Dependencies = append(targetSpec.Dependencies, &targetDependency)
	// Done
	return 0
}

// LUA function: goDependency
// Parameter:
//	name 				The target name. A string. Required.
// 	packages... 		The go packages
func luaGoDependency(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Get name
	name := l.ToString(1)
	if name == "" {
		panic("Require target name.")
	}
	targetSpec := buildSpec.GetTarget(name)
	if targetSpec == nil {
		panic("Target must be declared before go dependency")
	}
	// Get packages
	index := 1
	var packages []string
	for {
		index++
		pkg := l.ToString(index)
		if pkg == "" {
			break
		}
		packages = append(packages, pkg)
	}
	// Add
	targetSpec.GoDependencies = append(targetSpec.GoDependencies, &GoDependencySpec{
		Packages: packages,
	})
	// Done
	return 0
}

func luaPipDependency(ctx *ParseContext, l *lua.LState) int {
	buildSpec := ctx.BuildSpec()
	// Get name
	name := l.ToString(1)
	if name == "" {
		panic("Require target name.")
	}
	targetSpec := buildSpec.GetTarget(name)
	if targetSpec == nil {
		panic("Target must be declared before pip dependency")
	}
	// Get modules
	index := 1
	var modules []string
	for {
		index++
		pkg := l.ToString(index)
		if pkg == "" {
			break
		}
		modules = append(modules, pkg)
	}
	// Add
	targetSpec.PipDependencies = append(targetSpec.PipDependencies, &PipDependencySpec{
		Modules: modules,
	})
	// Done
	return 0

}

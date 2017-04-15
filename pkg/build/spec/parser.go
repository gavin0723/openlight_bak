// Author: lipixun
// Created Time : 日  3/12 15:34:23 2017
//
// File Name: parser.go
// Description:
//
//	The spec parser

package spec

import (
	"errors"

	"github.com/yuin/gopher-lua"
)

// ParseContext The parse context
type ParseContext struct {
	// The lua state
	luaState *lua.LState
	// The build data model
	buildSpec *BuildSpec
}

// NewParseContext Create a new parse context
func NewParseContext(safe bool, luaFunctions map[string]LUAFunctionProtoType) (*ParseContext, error) {
	if luaFunctions == nil {
		return nil, errors.New("Require lua functions")
	}
	// Create lua state
	var luaOpts []lua.Options
	if safe {
		// We do allow built in libraries in safe mode
		luaOpts = append(luaOpts, lua.Options{SkipOpenLibs: false})
	}
	luaState := lua.NewState(luaOpts...)
	// Create context
	ctx := ParseContext{
		luaState:  luaState,
		buildSpec: new(BuildSpec),
	}
	// Create lua functions
	for name, funcProtoType := range luaFunctions {
		if safe && !funcProtoType.Safe() {
			// Skip the unsafe functions
			continue
		}
		luaState.SetGlobal(name, luaState.NewFunction(funcProtoType.LUAFunction(&ctx)))
	}
	// Done
	return &ctx, nil
}

// ParseFile The file to parse
func (ctx *ParseContext) ParseFile(filename string) error {
	return ctx.luaState.DoFile(filename)
}

// ParseString The string to parse
func (ctx *ParseContext) ParseString(content string) error {
	return ctx.luaState.DoString(content)
}

// BuildSpec Get the build data model
func (ctx *ParseContext) BuildSpec() *BuildSpec {
	return ctx.buildSpec
}

// Close Close the parse context
func (ctx *ParseContext) Close() {
	ctx.luaState.Close()
}

// ParseFile Parse a build spec file
func ParseFile(filename string) (*ParseContext, error) {
	ctx, err := NewParseContext(true, LUAFunctions)
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	err = ctx.ParseFile(filename)
	return ctx, err
}

// ParseString Parse a build spec string
func ParseString(content string) (*ParseContext, error) {
	ctx, err := NewParseContext(true, LUAFunctions)
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	err = ctx.ParseString(content)
	return ctx, err
}

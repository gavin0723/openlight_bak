// Author: lipixun
// File Name: module.go
// Description:

package engine

import (
	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// Module defines the module interface
type Module interface {
	// LUAName returns the module name in lua
	LUAName() string
	// InitLUAModule initializes module into lua
	InitLUAModule(L *lua.LState) int
	// SetRuleFiles sets the rule files by this module
	SetRuleFiles(ruleFile *pbSpec.RuleFiles) error
}

// ModuleFactory defines a module factory interface which is used to create a new module
type ModuleFactory interface {
	// Name returns the name of the created module (which can be used for other modules to get this one)
	Name() string
	// Create a new module
	Create(ctx *Context) (Module, error)
}

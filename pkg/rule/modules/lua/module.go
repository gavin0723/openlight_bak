// Author: lipixun
// File Name: module.go
// Description:

package lua

import (
	"github.com/yuin/gopher-lua"
)

// ModuleContext defines the module context interface
type ModuleContext interface {
	// Modules return all modules
	Modules() []Module
	// GetModule returns the module
	GetModule(name string) Module
}

// Module defines the module interface
type Module interface {
	// Name returns the module name (in lua)
	Name() string
	// InitLInitLUAModule initializes (preload) module into lua
	InitLUAModule(L *lua.LState) int
}

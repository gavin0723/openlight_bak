// Author: lipixun
// File Name: context.go
// Description:

package engine

import (
	"github.com/yuin/gopher-lua"
)

// Context defines the engine runtime context
type Context struct {
	modules map[string]Module
	l       *lua.LState
}

// GetModule returns the module
func (ctx *Context) GetModule(name string) Module {
	return ctx.modules[name]
}

// Author: lipixun
// File Name: engine.go
// Description:

package engine

import (
	"github.com/ops-openlight/openlight/pkg/rule/common"
	"github.com/yuin/gopher-lua"
)

// Engine implements rule engien
type Engine struct {
	factories []ModuleFactory
}

// NewEngine creates a new Engine
func NewEngine(factories []ModuleFactory) *Engine {
	return &Engine{factories: factories}
}

// ParseFile parses a file
func (e *Engine) ParseFile(filename string) (*Context, error) {
	ctx, err := e.newContext()
	if err != nil {
		return nil, err
	}
	defer ctx.l.Close()
	if err := ctx.l.DoFile(filename); err != nil {
		return nil, err
	}
	return ctx, nil
}

// ParseString parses a string
func (e *Engine) ParseString(source string) (*Context, error) {
	ctx, err := e.newContext()
	if err != nil {
		return nil, err
	}
	defer ctx.l.Close()
	if err := ctx.l.DoString(source); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (e *Engine) newContext() (*Context, error) {
	L := lua.NewState(lua.Options{IncludeGoStackTrace: true})
	ctx := &Context{
		l:       L,
		modules: make(map[string]Module),
	}
	// Register common
	common.InitLUAGlobals(L)
	// Create modules by factory
	for _, factory := range e.factories {
		module, err := factory.Create(ctx)
		if err != nil {
			return nil, err
		}
		ctx.modules[factory.Name()] = module
		// Register
		L.PreloadModule(module.LUAName(), module.InitLUAModule)
	}
	// Done
	return ctx, nil
}

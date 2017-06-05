// Author: lipixun
// File Name: context.go
// Description:

package engine

import (
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
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

// GetRule returns the spec.RuleFiles object
func (ctx *Context) GetRule() (*pbSpec.RuleFiles, error) {
	var ruleFile pbSpec.RuleFiles
	for name, module := range ctx.modules {
		if err := module.SetRuleFiles(&ruleFile); err != nil {
			return nil, fmt.Errorf("Failed to set rule file of module [%v]: %v", name, err)
		}
	}
	return &ruleFile, nil
}

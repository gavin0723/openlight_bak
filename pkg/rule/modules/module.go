// Author: lipixun
// File Name: module.go
// Description:

package modules

import (
	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"

	"github.com/ops-openlight/openlight/pkg/rule/modules/build"
	"github.com/ops-openlight/openlight/pkg/rule/modules/runner"
)

// NewModules create moduels
func NewModules(ctx LUA.ModuleContext) map[string]LUA.Module {
	var modules = make(map[string]LUA.Module)
	// Create modules
	modules[build.LUANameModule] = build.NewModule(ctx)
	modules[runner.LUANameModule] = runner.NewModule(ctx)
	// Done
	return modules
}

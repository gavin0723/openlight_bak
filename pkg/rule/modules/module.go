// Author: lipixun
// File Name: module.go
// Description:

package modules

import (
	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"

	"github.com/ops-openlight/openlight/pkg/rule/modules/build"
)

// NewModules create moduels
func NewModules(ctx LUA.ModuleContext) map[string]LUA.Module {
	var modules = make(map[string]LUA.Module)
	// Create modules
	modules[build.LUANameModule] = build.NewModule(ctx)
	// Done
	return modules
}

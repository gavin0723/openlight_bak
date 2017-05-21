// Author: lipixun
// File Name: loader.go
// Description:

package rule

import (
	"github.com/ops-openlight/openlight/pkg/rule/engine"
	"github.com/ops-openlight/openlight/pkg/rule/modules/build"
	"github.com/ops-openlight/openlight/pkg/rule/modules/runner"
)

// NewEngine creates a new default engine
func NewEngine() *engine.Engine {
	return engine.NewEngine([]engine.ModuleFactory{
		build.NewModuleFactory(),
		runner.NewModuleFactory(),
	})
}

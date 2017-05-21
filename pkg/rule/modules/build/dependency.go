// Author: lipixun
// File Name: dependency.go
// Description:

package build

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/rule/common"
)

// Ensure the interface is implemented
var _ common.Object = (*Dependency)(nil)

const (
	targetDependencyLUAName     = "TargetDependency"
	targetDependencyLUATypeName = "build-dependency-target"
	goDependencyLUAName         = "GoDependency"
	goDependencyLUATypeName     = "build-dependency-go"
	pipDependencyLUAName        = "PipDependency"
	pipDependencyLUATypeName    = "build-dependency-pip"
)

// registerTargetDependencyType registers target dependency type in lua
func registerTargetDependencyType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(targetDependencyLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewTargetDependencyFromLUA))
	L.SetField(mod, targetDependencyLUAName, mt)
}

// registerGoDependencyType registers target dependency type in lua
func registerGoDependencyType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(goDependencyLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewGoDependencyFromLUA))
	L.SetField(mod, goDependencyLUAName, mt)
}

// registerPipDependencyType registers target dependency type in lua
func registerPipDependencyType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(pipDependencyLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewPipDependencyFromLUA))
	L.SetField(mod, pipDependencyLUAName, mt)
}

// Dependency defines the dependency
type Dependency pbSpec.Dependency

// NewTargetDependencyFromLUA creates a new TargetDependency from LUA
func NewTargetDependencyFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	reference, err := params.GetString("reference")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [reference]: %v", err)
	}
	target, err := params.GetString("target")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [target]: %v", err)
	}
	if target == "" {
		return nil, errors.New("Require target")
	}
	build, err := params.GetBool("build")
	// Create a new dependency
	dep := &Dependency{
		Dependency: &pbSpec.Dependency_Target{
			Target: &pbSpec.TargetDependency{
				Reference: reference,
				Target:    target,
				Build:     build,
			},
		},
	}
	// Done
	return dep.GetLUAUserData(L), nil
}

// NewGoDependencyFromLUA creates a new GoDependency from LUA
func NewGoDependencyFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	pkg, err := params.GetString("package")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [package]: %v", err)
	}
	if pkg == "" {
		return nil, errors.New("Require package")
	}
	// Create a new dependency
	dep := &Dependency{
		Dependency: &pbSpec.Dependency_Go{
			Go: &pbSpec.GoDependency{
				Package: pkg,
			},
		},
	}
	// Done
	return dep.GetLUAUserData(L), nil
}

// NewPipDependencyFromLUA creates a new PipDependency from LUA
func NewPipDependencyFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	module, err := params.GetString("module")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [module]: %v", err)
	}
	if module == "" {
		return nil, errors.New("Require module")
	}
	// Create a new dependency
	dep := &Dependency{
		Dependency: &pbSpec.Dependency_Pip{
			Pip: &pbSpec.PipDependency{
				Module: module,
			},
		},
	}
	// Done
	return dep.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (dep *Dependency) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = dep
	// Done
	return ud
}

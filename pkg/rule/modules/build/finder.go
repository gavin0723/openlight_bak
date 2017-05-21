// Author: lipixun
// File Name: finder.go
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
var _ Finder = (*PythonFinder)(nil)
var _ Finder = (*GoFinder)(nil)

const (
	pythonFinderLUAName     = "PythonFinder"
	pythonFinderLUATypeName = "build-finder-python"
	goFinderLUAName         = "GoFinder"
	goFinderLUATypeName     = "build-finder-go"
)

// registerPythonFinderType registers python finder type in lua
func registerPythonFinderType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(pythonFinderLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewPythonFinderFromLUA))
	L.SetField(mod, pythonFinderLUAName, mt)
}

// registerGoFinderType registers go finder type in lua
func registerGoFinderType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(goFinderLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewGoFinderFromLUA))
	L.SetField(mod, goFinderLUAName, mt)
}

// Finder represents the local finder
type Finder interface {
	common.Object
	// GetProto returns the protobuf object
	GetProto(name string) *pbSpec.Finder
}

// PythonFinder implements the python finder
type PythonFinder pbSpec.PythonFinder

// NewPythonFinder creates a new PythonFinder
func NewPythonFinder(module string, parent int) *PythonFinder {
	return &PythonFinder{Module: module, Parent: int32(parent)}
}

// NewPythonFinderFromLUA creates a new PythonFinder from lua
func NewPythonFinderFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	module, err := params.GetString("module")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [module]: %v", err)
	}
	if module == "" {
		return nil, errors.New("Require module")
	}
	parent, err := params.GetInt("parent")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [parent]: %v", err)
	}
	if parent < 0 {
		return nil, errors.New("Parent must be a positive number")
	}
	finder := NewPythonFinder(module, parent)
	// Done
	return finder.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (finder *PythonFinder) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = finder
	L.SetMetatable(ud, L.GetTypeMetatable(pythonFinderLUATypeName))
	// Done
	return ud
}

// GetProto returns the protobuf object
func (finder *PythonFinder) GetProto(name string) *pbSpec.Finder {
	return &pbSpec.Finder{
		Name: name,
		Finder: &pbSpec.Finder_Python{
			Python: (*pbSpec.PythonFinder)(finder),
		},
	}
}

// GoFinder implements the go finder
type GoFinder pbSpec.GoFinder

// NewGoFinder creates a new GoFinder
func NewGoFinder(pkg, root string, parent int) *GoFinder {
	return &GoFinder{Package: pkg, Root: root, Parent: int32(parent)}
}

// NewGoFinderFromLUA creates a new GoFinder from lua
func NewGoFinderFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	pkg, err := params.GetString("package")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [package]: %v", err)
	}
	if pkg == "" {
		return nil, errors.New("Require package")
	}
	root, err := params.GetString("root")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [root]: %v", err)
	}
	parent, err := params.GetInt("parent")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [parent]: %v", err)
	}
	if parent < 0 {
		return nil, errors.New("Parent must be a positive number")
	}
	finder := NewGoFinder(pkg, root, parent)
	// Done
	return finder.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (finder *GoFinder) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = finder
	L.SetMetatable(ud, L.GetTypeMetatable(goFinderLUATypeName))
	// Done
	return ud
}

// GetProto returns the protobuf object
func (finder *GoFinder) GetProto(name string) *pbSpec.Finder {
	return &pbSpec.Finder{
		Name: name,
		Finder: &pbSpec.Finder_Go{
			Go: (*pbSpec.GoFinder)(finder),
		},
	}
}

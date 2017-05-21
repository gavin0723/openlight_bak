// Author: lipixun
// File Name: reference.go
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
var _ common.Object = (*Reference)(nil)

const (
	referenceLUAName     = "Reference"
	referenceLUATypeName = "Build-Reference"
)

// registerReferenceType registers reference type in lua
func registerReferenceType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(referenceLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewReferenceFromLUA))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"finders": luaFuncReferenceFinders,
	}))
	// Set field
	L.SetField(mod, referenceLUAName, mt)
}

// Reference represents the reference of build package
type Reference struct {
	Remote  string
	Finders *common.NamedCollection
	Path    string
}

// NewReference creates a new Reference
func NewReference(remote, path string) *Reference {
	return &Reference{
		Remote: remote,
		Path:   path,
	}
}

// NewReferenceFromLUA creates a new Reference from lua
func NewReferenceFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	remote, err := params.GetString("remote")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [remote]: %v", err)
	}
	path, err := params.GetString("path")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [path]: %v", err)
	}
	if remote == "" && path == "" {
		return nil, errors.New("Please specify remote or path at least")
	}
	ref := NewReference(remote, path)
	// Done
	return ref.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (ref *Reference) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = ref
	L.SetMetatable(ud, L.GetTypeMetatable(referenceLUATypeName))
	// Done
	return ud
}

// GetProto returns the protobuf object
func (ref *Reference) GetProto() *pbSpec.Reference {
	var pbReference pbSpec.Reference
	pbReference.Remote = ref.Remote
	pbReference.Path = ref.Path
	// Finders
	if ref.Finders != nil {
		for _, item := range ref.Finders.Items() {
			pbReference.Finders = append(pbReference.Finders, item.Value.(Finder).GetProto(item.Name))
		}
	}
	// Done
	return &pbReference
}

func convertFinder(value lua.LValue) (common.Object, error) {
	if value == nil {
		return nil, errors.New("Invalid nil value")
	}
	if value.Type() != lua.LTUserData {
		return nil, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTUserData, value.Type())
	}
	finder, ok := value.(*lua.LUserData).Value.(Finder)
	if !ok {
		return nil, errors.New("Expect Finder object")
	}
	return finder, nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// luaReferenceSelf get lua reference self
func luaReferenceSelf(L *lua.LState) *Reference {
	ud := L.CheckUserData(1)
	if ref, ok := ud.Value.(*Reference); ok {
		return ref
	}
	L.ArgError(1, "Reference expected")
	return nil
}

// luaFuncReferenceFinders defines Reference.finders function in lua
func luaFuncReferenceFinders(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.ArgError(0, "Invalid arguments")
		return 0
	}
	ref := luaReferenceSelf(L)
	if ref.Finders == nil {
		ref.Finders = common.NewNamedCollection(convertFinder)
	}
	// Return name
	L.Push(ref.Finders.GetLUAUserData(L))
	return 1
}

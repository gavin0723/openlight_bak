// Author: lipixun
// File Name: reference.go
// Description:

package build

import (
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

// Exposed lua infos
const (
	LUATypeReference = "Build-Reference"
)

// ReferenceLUAFuncs defines all lua functions for reference
var ReferenceLUAFuncs = map[string]lua.LGFunction{
	"name":              LUAFuncReferenceName,
	"remote":            LUAFuncReferenceRemote,
	"options":           LUA.FuncObjectOptions,
	"localFinder":       LUAFuncReferenceLocalFinder,
	"pythonLocalFinder": LUAFuncReferencePythonLocalFinder,
}

// RegisterReferenceType registers reference type in lua
func RegisterReferenceType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypeReference)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), ReferenceLUAFuncs))
}

// Reference represents the reference of build package
type Reference struct {
	LUA.Object
	Name         string
	Remote       string
	LocalFinders []LocalFinder
}

// NewReference creates a new Reference
func NewReference(name, remote string, localFinders []LocalFinder, options *lua.LTable) *Reference {
	var ref = Reference{
		Name:         name,
		Remote:       remote,
		LocalFinders: localFinders,
	}
	ref.Object = LUA.NewObject(LUATypeReference, options, &ref)
	// Done
	return &ref
}

// GetProto returns the protobuf object
func (ref *Reference) GetProto() (*pbSpec.Reference, error) {
	var pbReference pbSpec.Reference
	pbReference.Name = ref.Name
	pbReference.Remote = ref.Remote
	for _, finder := range ref.LocalFinders {
		pbFinder, err := finder.GetProto()
		if err != nil {
			return nil, err
		}
		pbReference.LocalFinders = append(pbReference.LocalFinders, pbFinder)
	}
	// Done
	return &pbReference, nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUAReferenceSelf get lua reference self
func LUAReferenceSelf(L *lua.LState) *Reference {
	ud := L.CheckUserData(1)
	if ref, ok := ud.Value.(*Reference); ok {
		return ref
	}
	L.ArgError(1, "Reference expected")
	return nil
}

// LUAFuncReferenceName defines package.name function in lua
func LUAFuncReferenceName(L *lua.LState) int {
	ref := LUAReferenceSelf(L)
	if ref == nil {
		return 0
	}
	if L.GetTop() != 1 {
		L.ArgError(0, "Invalid arguments")
		return 0
	}
	// Return name
	L.Push(lua.LString(ref.Name))
	return 1
}

// LUAFuncReferenceRemote defines package.remote function in lua
func LUAFuncReferenceRemote(L *lua.LState) int {
	ref := LUAReferenceSelf(L)
	if ref == nil {
		return 0
	}
	if L.GetTop() == 1 {
		// Get
		L.Push(lua.LString(ref.Remote))
		return 1
	} else if L.GetTop() == 2 {
		// Set
		ref.Remote = L.CheckString(2)
		return 0
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

// LUAFuncReferenceLocalFinder defines reference.localFinder function in lua
func LUAFuncReferenceLocalFinder(L *lua.LState) int {
	ref := LUAReferenceSelf(L)
	if ref == nil {
		return 0
	}
	if L.GetTop() != 2 {
		L.ArgError(0, "Invalid arguments")
		return 0
	}
	name := L.CheckString(2)
	for _, finder := range ref.LocalFinders {
		if finder.GetName() == name {
			L.Push(finder.GetLUAUserData(L))
			return 1
		}
	}
	return 0
}

// LUAFuncReferencePythonLocalFinder defines reference.pythonLocalFinder function in lua
func LUAFuncReferencePythonLocalFinder(L *lua.LState) int {
	ref := LUAReferenceSelf(L)
	if ref == nil {
		return 0
	}
	// Create python finder
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	for _, finder := range ref.LocalFinders {
		if finder.GetName() == name {
			L.ArgError(2, fmt.Sprintf("Duplicated finder name [%v]", name))
			return 0
		}
	}
	module := L.CheckString(3)
	if module == "" {
		L.ArgError(3, "Require module")
		return 0
	}
	finder := NewPythonLocalFinder(name, module, L.ToTable(4))
	ref.LocalFinders = append(ref.LocalFinders, finder)
	// Return
	L.Push(finder.GetLUAUserData(L))
	return 1
}

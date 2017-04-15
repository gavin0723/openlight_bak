// Author: lipixun
// File Name: object.go
// Description:

package lua

import (
	"github.com/yuin/gopher-lua"
)

// Object represents the common lua base object
type Object interface {
	// GetOptions returns the options
	GetOptions() *lua.LTable
	// SetOptions sets options
	SetOptions(options *lua.LTable)
	// GetLUAUserData returns the lua user data
	GetLUAUserData(L *lua.LState) *lua.LUserData
}

// _Object implements Object interface
type _Object struct {
	luaTypeName string
	options     *lua.LTable
	object      interface{}
}

// NewObject creates a new Object
func NewObject(luaTypeName string, options *lua.LTable, object interface{}) Object {
	return &_Object{
		luaTypeName: luaTypeName,
		options:     options,
		object:      object,
	}
}

// GetOptions returns the options
func (o *_Object) GetOptions() *lua.LTable {
	return o.options
}

// SetOptions sets options
func (o *_Object) SetOptions(options *lua.LTable) {
	o.options = options
}

// GetLUAUserData returns the lua user data
func (o *_Object) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = o.object
	if o.options == nil {
		o.options = L.NewTable()
	}
	L.SetMetatable(ud, L.GetTypeMetatable(o.luaTypeName))
	// Done
	return ud
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// FuncObjectSelf get lua object self
func FuncObjectSelf(L *lua.LState) Object {
	ud := L.CheckUserData(1)
	if obj, ok := ud.Value.(Object); ok {
		return obj
	}
	L.ArgError(1, "Object expected")
	return nil
}

// FuncObjectOptions defines object.options in lua
func FuncObjectOptions(L *lua.LState) int {
	obj := FuncObjectSelf(L)
	if obj == nil {
		return 0
	}
	if L.GetTop() == 1 {
		// Get options
		L.Push(obj.GetOptions())
		return 1
	} else if L.GetTop() == 2 {
		// Set options
		options := L.CheckTable(2)
		if options == nil {
			return 0
		}
		obj.SetOptions(options)
		return 0
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

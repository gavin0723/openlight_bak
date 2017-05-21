// Author: lipixun
// File Name: helper.go
// Description:

package common

import (
	"github.com/yuin/gopher-lua"
)

// InitLUAGlobals initializes lua global
func InitLUAGlobals(L *lua.LState) {
	// Register types
	registerNamedCollectionType(L)
	registerOptionsType(L)
}

// LUANewObjectFunc defines a fucntion of go new
type LUANewObjectFunc func(*lua.LState, Parameters) (lua.LValue, error)

// NewLUANewObjectFunc defines RunCommand.new function in lua
func NewLUANewObjectFunc(newfunc LUANewObjectFunc) lua.LGFunction {
	if newfunc == nil {
		panic("Require new function")
	}
	// Return the new function
	return func(L *lua.LState) int {
		if L.GetTop() > 1 {
			L.ArgError(2, "Too many arguments")
		}
		var err error
		var params Parameters
		if L.GetTop() == 1 {
			// Get table and convert to parameters
			params, err = NewParametersFromLUATable(L.CheckTable(1))
			if err != nil {
				L.ArgError(0, err.Error())
			}
		}
		obj, err := newfunc(L, params)
		if err != nil {
			L.ArgError(0, err.Error())
		}
		// Return the object
		L.Push(obj)
		return 1
	}
}

// NewLUANewObjectFunction creates a new lua function for creating new lua object
func NewLUANewObjectFunction(L *lua.LState, newfunc LUANewObjectFunc) *lua.LFunction {
	return L.NewFunction(NewLUANewObjectFunc(newfunc))
}

// Author: lipixun
// File Name: object.go
// Description:

package common

import (
	"github.com/yuin/gopher-lua"
)

// Object defines the interface of all objects which are used in rule
type Object interface {
	// GetLUAUserData returns the lua user data
	GetLUAUserData(L *lua.LState) *lua.LUserData
}

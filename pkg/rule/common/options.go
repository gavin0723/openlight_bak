// Author: lipixun
// File Name: options.go
// Description:

package common

import (
	"fmt"

	"github.com/yuin/gopher-lua"
)

// Ensure the interface is implemented
var _ Object = (*Options)(nil)

const (
	optionsLUATypeName = "common-options"
	optionsLUAName     = "Options"
)

// OptionsValueConvertFunc represents the options value convert function
type OptionsValueConvertFunc func(key string, value lua.LValue) (interface{}, error)

// Options represents the options
type Options struct {
	values      map[string]*OptionValue
	convertFunc OptionsValueConvertFunc
}

// OptionValue represents the option value
type OptionValue struct {
	value    interface{}
	luaValue lua.LValue
}

// Value returns the value
func (v *OptionValue) Value() interface{} {
	return v.value
}

// registerOptionsType registers Options type
func registerOptionsType(L *lua.LState) {
	mt := L.NewTypeMetatable(optionsLUATypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":    luaOptionsGet,
		"set":    luaOptionsSet,
		"delete": luaOptionsDelete,
	}))
	// Add type to global
	L.SetGlobal(optionsLUAName, mt)
}

// NewOptions creates a new Options
func NewOptions(convertFunc OptionsValueConvertFunc) *Options {
	if convertFunc == nil {
		panic("Require convert function")
	}
	return &Options{convertFunc: convertFunc, values: make(map[string]*OptionValue)}
}

// Get a value
func (c *Options) Get(name string) *OptionValue {
	return c.values[name]
}

// Add a value
func (c *Options) Add(name string, value lua.LValue) error {
	if item := c.values[name]; item != nil {
		return fmt.Errorf("Name [%v] has already be added", name)
	}
	convertedValue, err := c.convertFunc(name, value)
	if err != nil {
		return err
	}
	optionValue := &OptionValue{value: convertedValue, luaValue: value}
	c.values[name] = optionValue
	// Done
	return nil
}

// Delete a value
func (c *Options) Delete(name string) {
	delete(c.values, name)
}

// Keys return all keys
func (c *Options) Keys() []string {
	var keys []string
	for key := range c.values {
		keys = append(keys, key)
	}
	return keys
}

// GetLUAUserData returns the lua user data
func (c *Options) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = c
	L.SetMetatable(ud, L.GetTypeMetatable(optionsLUATypeName))
	// Done
	return ud
}

// luaOptionsSelf returns the named collection
func luaOptionsSelf(L *lua.LState) *Options {
	ud := L.CheckUserData(1)
	if obj, ok := ud.Value.(*Options); ok {
		return obj
	}
	L.ArgError(1, "Not a Options object")
	return nil
}

func luaOptionsGet(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require name")
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many names")
	}
	// Get value and return
	c := luaOptionsSelf(L)
	value := c.Get(L.CheckString(2))
	if value != nil {
		L.Push(value.luaValue)
		return 1
	}
	return 0
}

func luaOptionsSet(L *lua.LState) int {
	if L.GetTop() != 3 {
		L.ArgError(3, "Invalid arguments")
	}
	// Add
	c := luaOptionsSelf(L)
	if err := c.Add(L.CheckString(2), L.Get(3)); err != nil {
		L.ArgError(1, err.Error())
	}
	// Done
	return 0
}

func luaOptionsDelete(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require name")
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many names")
	}
	// Delete name
	c := luaOptionsSelf(L)
	c.Delete(L.CheckString(2))
	// Done
	return 0
}

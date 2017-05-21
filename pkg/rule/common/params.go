// Author: lipixun
// File Name: params.go
// Description:

package common

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"
)

// Parameters represents a key-value parameter map from lua
type Parameters map[string]lua.LValue

// NewParameters creates a new Parameters
func NewParameters() Parameters {
	return make(Parameters)
}

// NewParametersFromLUATable creates a new Parameters from lua table
func NewParametersFromLUATable(table *lua.LTable) (Parameters, error) {
	var err error
	params := NewParameters()
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if err == nil {
			if key.Type() != lua.LTString {
				err = errors.New("Parameter name must be a string")
				return
			}
			params[string(key.(lua.LString))] = value
		}
	})
	if err != nil {
		return nil, err
	}
	// Done
	return params, nil
}

// GetString returns the string of the key
func (params Parameters) GetString(key string) (string, error) {
	value := params[key]
	if value == nil {
		return "", nil
	}
	if value.Type() != lua.LTString {
		return "", fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTString, value.Type())
	}
	return string(value.(lua.LString)), nil
}

// GetStringSlice returns the string slice of the key
func (params Parameters) GetStringSlice(key string) ([]string, error) {
	value := params[key]
	if value == nil {
		return nil, nil
	}
	if value.Type() != lua.LTTable {
		return nil, errors.New("Not a string list")
	}
	return ConvertTableToStringSlice(value.(*lua.LTable))
}

// GetInt returns the int of the key
func (params Parameters) GetInt(key string) (int, error) {
	value := params[key]
	if value == nil {
		return 0, nil
	}
	if value.Type() != lua.LTNumber {
		return 0, errors.New("Not a number")
	}
	return int(value.(lua.LNumber)), nil
}

// GetFloat64 returns the float64 of the key
func (params Parameters) GetFloat64(key string) (float64, error) {
	value := params[key]
	if value == nil {
		return 0, nil
	}
	if value.Type() != lua.LTNumber {
		return 0, errors.New("Not a number")
	}
	return float64(value.(lua.LNumber)), nil
}

// GetBool returns the bool of the key
func (params Parameters) GetBool(key string) (bool, error) {
	value := params[key]
	if value == nil {
		return false, nil
	}
	if value.Type() != lua.LTBool {
		return false, errors.New("Not a bool")
	}
	return bool(value.(lua.LBool)), nil
}

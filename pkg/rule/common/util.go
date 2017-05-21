// Author: lipixun
// File Name: util.go
// Description:

package common

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"
)

// GetIntFromTable gets int from table
func GetIntFromTable(table *lua.LTable, key string) (int, error) {
	value := table.RawGetString(key)
	if value.Type() != lua.LTNumber {
		return 0, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTNumber, value.Type())
	}
	// NOTE: Here we doesn't check int or float
	return int(value.(lua.LNumber)), nil
}

// GetBoolFromTable gets bool from table
func GetBoolFromTable(table *lua.LTable, key string) (bool, error) {
	value := table.RawGetString(key)
	if value.Type() != lua.LTBool {
		return false, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTBool, value.Type())
	}
	return bool(value.(lua.LBool)), nil
}

// TryGetBoolFromTable gets bool from table with default
func TryGetBoolFromTable(table *lua.LTable, key string, defaultValue bool) (bool, error) {
	value := table.RawGetString(key)
	if value == lua.LNil {
		return defaultValue, nil
	}
	if value.Type() != lua.LTBool {
		return false, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTBool, value.Type())
	}
	return bool(value.(lua.LBool)), nil
}

// GetStringFromTable gets string from table
func GetStringFromTable(table *lua.LTable, key string) (string, error) {
	value := table.RawGetString(key)
	if value.Type() != lua.LTString {
		return "", fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTString, value.Type())
	}
	return string(value.(lua.LString)), nil
}

// TryGetStringFromTable gets string from table
func TryGetStringFromTable(table *lua.LTable, key string, defaultValue string) (string, error) {
	value := table.RawGetString(key)
	if value == lua.LNil {
		return defaultValue, nil
	}
	if value.Type() != lua.LTString {
		return "", fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTString, value.Type())
	}
	return string(value.(lua.LString)), nil
}

// GetStringSliceFromTable gets string slice from table
func GetStringSliceFromTable(table *lua.LTable, key string) ([]string, error) {
	value := table.RawGetString(key)
	if value.Type() != lua.LTTable {
		return nil, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTTable, value.Type())
	}
	return ConvertTableToStringSlice(value.(*lua.LTable))
}

// TryGetStringSliceFromTable gets string slice from table
func TryGetStringSliceFromTable(table *lua.LTable, key string, defaultValue []string) ([]string, error) {
	value := table.RawGetString(key)
	if value == lua.LNil {
		return defaultValue, nil
	}
	return GetStringSliceFromTable(table, key)
}

// ConvertTableToStringSlice converts table to string slice
func ConvertTableToStringSlice(table *lua.LTable) ([]string, error) {
	if table == nil {
		return nil, nil
	}
	var strs []string
	for i := 1; i <= table.Len(); i++ {
		value := table.RawGetInt(i)
		if value == lua.LNil {
			return nil, errors.New("Not a string list")
		} else if value.Type() != lua.LTString {
			return nil, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTString, value.Type())
		}
		strs = append(strs, string(value.(lua.LString)))
	}
	return strs, nil
}

// ConvertValueToStringSlice converts a lua value to string slice
func ConvertValueToStringSlice(value lua.LValue) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	if value.Type() != lua.LTTable {
		return nil, fmt.Errorf("Expect [%v] type, actually got [%v] type", lua.LTTable, value.Type())
	}
	return ConvertTableToStringSlice(value.(*lua.LTable))
}

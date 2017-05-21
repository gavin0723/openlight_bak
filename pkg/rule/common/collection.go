// Author: lipixun
// File Name: collection.go
// Description:
//
//	Collection implementation
//

package common

import (
	"fmt"

	"github.com/yuin/gopher-lua"
)

// Ensure the interface is implemented
var _ Object = (*NamedCollection)(nil)

const (
	namedCollectionLUATypeName = "common-namedcollection"
	namedCollectionLUAName     = "NamedCollection"
)

// CollectionValueConvertFunc defines a function which is used to validate the value to be added into collection
type CollectionValueConvertFunc func(value lua.LValue) (Object, error)

// NamedCollection implements a common named collection data type in lua
type NamedCollection struct {
	items       []*NamedCollectionItem
	itemMap     map[string]*NamedCollectionItem
	convertFunc CollectionValueConvertFunc
}

// NamedCollectionItem defineds the named item
type NamedCollectionItem struct {
	Name  string
	Value Object
}

// registerNamedCollectionType registers NamedCollection type
func registerNamedCollectionType(L *lua.LState) {
	mt := L.NewTypeMetatable(namedCollectionLUATypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":    luaNamedCollectionGet,
		"add":    luaNamedCollectionAdd,
		"delete": luaNamedCollectionDelete,
	}))
	// Add type to global
	L.SetGlobal(namedCollectionLUAName, mt)
}

// NewNamedCollection creates a new NamedCollection
func NewNamedCollection(convertFunc CollectionValueConvertFunc) *NamedCollection {
	if convertFunc == nil {
		panic("Require convert function")
	}
	return &NamedCollection{convertFunc: convertFunc, itemMap: make(map[string]*NamedCollectionItem)}
}

// Get a value
func (c *NamedCollection) Get(name string) Object {
	if item := c.itemMap[name]; item != nil {
		return item.Value
	}
	return nil
}

// Add a value
func (c *NamedCollection) Add(name string, value lua.LValue) error {
	if item := c.itemMap[name]; item != nil {
		return fmt.Errorf("Name [%v] has already be added", name)
	}
	obj, err := c.convertFunc(value)
	if err != nil {
		return err
	}
	item := &NamedCollectionItem{Name: name, Value: obj}
	c.itemMap[name] = item
	c.items = append(c.items, item)
	// Done
	return nil
}

// Delete a value
func (c *NamedCollection) Delete(name string) {
	if item := c.itemMap[name]; item != nil {
		delete(c.itemMap, name)
		var items []*NamedCollectionItem
		for _, item := range c.items {
			if item.Name != name {
				items = append(items, item)
			}
		}
		c.items = items
	}
}

// Items return all items
func (c *NamedCollection) Items() []*NamedCollectionItem {
	return c.items
}

// GetLUAUserData returns the lua user data
func (c *NamedCollection) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = c
	L.SetMetatable(ud, L.GetTypeMetatable(namedCollectionLUATypeName))
	// Done
	return ud
}

// luaNamedCollectionSelf returns the named collection
func luaNamedCollectionSelf(L *lua.LState) *NamedCollection {
	ud := L.CheckUserData(1)
	if obj, ok := ud.Value.(*NamedCollection); ok {
		return obj
	}
	L.ArgError(1, "Not a NamedCollection object")
	return nil
}

func luaNamedCollectionGet(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require name")
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many names")
	}
	// Get value and return
	c := luaNamedCollectionSelf(L)
	value := c.Get(L.CheckString(2))
	if value != nil {
		L.Push(value.GetLUAUserData(L))
		return 1
	}
	return 0
}

func luaNamedCollectionAdd(L *lua.LState) int {
	if L.GetTop() != 3 {
		L.ArgError(3, "Invalid arguments")
	}
	// Add
	c := luaNamedCollectionSelf(L)
	if err := c.Add(L.CheckString(2), L.Get(3)); err != nil {
		L.ArgError(1, err.Error())
	}
	// Done
	return 0
}

func luaNamedCollectionDelete(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require name")
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many names")
	}
	// Delete name
	c := luaNamedCollectionSelf(L)
	c.Delete(L.CheckString(2))
	// Done
	return 0
}

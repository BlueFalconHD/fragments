package main

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type LFragment struct {
	Fragment   *Fragment
	Parent     *LFragment
	LocalMeta  *CoreTable
	GlobalMeta *CoreTable
}

func (f *LFragment) getFragmentTree() []*LFragment {
	var tree []*LFragment
	current := f
	for current != nil {
		tree = append([]*LFragment{current}, tree...)
		current = current.Parent
	}
	return tree
}

const luaFragmentTypeName = "fragment"

func registerFragmentType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaFragmentTypeName)
	L.SetGlobal("fragment", mt)
	// Static attributes
	L.SetField(mt, "new", L.NewFunction(newFragment))
	// Methods and metamethods
	L.SetField(mt, "__index", L.NewFunction(fragmentIndex))
	L.SetField(mt, "__newindex", L.NewFunction(fragmentNewIndex))
}

func newFragment(L *lua.LState) int {
	var parent *LFragment
	if L.GetTop() > 0 {
		// Optional parent fragment passed as an argument
		ud := L.CheckUserData(1)
		if v, ok := ud.Value.(*LFragment); ok {
			parent = v
		} else {
			L.ArgError(1, "fragment expected as parent")
			return 0
		}
	}
	f := &LFragment{
		Fragment:   nil,
		Parent:     parent,
		LocalMeta:  NewCoreTable(make(map[string]CoreType)),
		GlobalMeta: NewCoreTable(make(map[string]CoreType)),
	}
	ud := L.NewUserData()
	ud.Value = f
	L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
	L.Push(ud)
	return 1
}

func checkFragment(L *lua.LState) *LFragment {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*LFragment); ok {
		return v
	}
	L.ArgError(1, "fragment expected")
	return nil
}

var fragmentMethods = map[string]lua.LGFunction{
	"getMeta":       fragmentGetMeta,
	"getGlobalMeta": fragmentGetGlobalMeta,
	"setMeta":       fragmentSetMeta,
	"setGlobalMeta": fragmentSetGlobalMeta,
	"parent":        fragmentParent,
}

func fragmentIndex(L *lua.LState) int {
	f := checkFragment(L)
	field := L.CheckString(2)

	// First, check if field is a method
	if method, ok := fragmentMethods[field]; ok {
		L.Push(L.NewFunction(method))
		return 1
	}

	// Handle properties
	switch field {
	case "name":
		L.Push(lua.LString(f.Fragment.name))
	case "code":
		L.Push(lua.LString(f.Fragment.code))
	case "fspath":
		L.Push(lua.LString(f.Fragment.fspath))
	case "site":
		L.ArgError(2, "site is a custom type that hasn't been implemented yet")
	case "options":
		L.ArgError(2, "options is a custom type that hasn't been implemented yet")
	default:
		L.Push(lua.LNil)
	}
	return 1
}

func fragmentNewIndex(L *lua.LState) int {
	f := checkFragment(L)
	field := L.CheckString(2)
	value := L.Get(3)

	switch field {
	case "localMeta":
		if value.Type() == lua.LTTable {
			f.LocalMeta = NewCoreTableL(value.(*lua.LTable))
		} else {
			L.ArgError(3, "table expected for localMeta")
		}
	case "globalMeta":
		if value.Type() == lua.LTTable {
			f.GlobalMeta = NewCoreTableL(value.(*lua.LTable))
		} else {
			L.ArgError(3, "table expected for globalMeta")
		}
	default:
		L.ArgError(2, "unexpected field: "+field)
	}
	return 0
}

func fragmentGetMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := getNestedValue(f.LocalMeta, key)
	L.Push(value.luaType(L))
	return 1
}

func fragmentGetGlobalMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := getNestedValue(f.GlobalMeta, key)
	L.Push(value.luaType(L))
	return 1
}

func fragmentSetMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := luaToCoreType(L.Get(3))
	setNestedValue(f.LocalMeta, key, value)
	return 0
}

func fragmentSetGlobalMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := luaToCoreType(L.Get(3))
	setNestedValue(f.GlobalMeta, key, value)
	return 0
}

func fragmentParent(L *lua.LState) int {
	f := checkFragment(L)
	if f.Parent != nil {
		ud := L.NewUserData()
		ud.Value = f.Parent
		L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
		L.Push(ud)
	} else {
		L.Push(lua.LNil)
	}
	return 1
}

// Helper functions to handle nested keys
func getNestedValue(table *CoreTable, key string) CoreType {
	keys := strings.Split(key, ".")
	current := table
	for i, k := range keys {
		val, ok := current.v[k]
		if !ok {
			return NewCoreNil()
		}
		if i == len(keys)-1 {
			return val
		}
		// Intermediate keys, expect CoreTable
		if ct, ok := val.(*CoreTable); ok {
			current = ct
		} else {
			return NewCoreNil()
		}
	}
	return NewCoreNil()
}

func setNestedValue(table *CoreTable, key string, value CoreType) {
	keys := strings.Split(key, ".")
	current := table
	for i, k := range keys {
		if i == len(keys)-1 {
			// Last key, set the value
			current.v[k] = value
			return
		}
		// Intermediate keys, expect CoreTable
		if val, ok := current.v[k]; ok {
			if ct, ok := val.(*CoreTable); ok {
				current = ct
			} else {
				// Not a table, create a new one
				newTable := NewCoreTable(make(map[string]CoreType))
				current.v[k] = newTable
				current = newTable
			}
		} else {
			// Key does not exist, create a new table
			newTable := NewCoreTable(make(map[string]CoreType))
			current.v[k] = newTable
			current = newTable
		}
	}
}

func testLua() {
	L := lua.NewState()
	defer L.Close()
	registerFragmentType(L)
	if err := L.DoString(`
		-- Creating parent fragment
		parent = fragment.new()
		parent:setMeta("title", "Parent Fragment")
		
		-- Creating child fragment with parent
		child = fragment.new(parent)
		child:setMeta("title", "Child Fragment")
		
		-- Accessing parent from child
		print("Child's title:", child:getMeta("title"))
		print("Parent's title via child:", child:parent():getMeta("title"))
		
		-- Checking if parent of parent is nil
		print("Parent's parent:", parent:parent())

	
	`); err != nil {
		panic(err)
	}

	// Retrieve the child Fragment object from Lua
	childUd := L.GetGlobal("child")
	luaChildUd, ok := childUd.(*lua.LUserData)
	if !ok {
		fmt.Println("child is not a userdata")
		return
	}

	childFragment, ok := luaChildUd.Value.(*LFragment)
	if !ok {
		fmt.Println("child ud.Value is not a *LFragment")
		return
	}

	// Access the meta data for the child fragment
	fmt.Println("Child Fragment title:", childFragment.LocalMeta.v["title"].goType())

	// Access the parent fragment from the child
	if childFragment.Parent != nil {
		parentFragment := childFragment.Parent
		fmt.Println("Parent Fragment title:", parentFragment.LocalMeta.v["title"].goType())
	} else {
		fmt.Println("Child fragment has no parent.")
	}
}

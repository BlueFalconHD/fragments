package main

import (
	"github.com/charmbracelet/log"
	libs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
	"strings"
)

type LFragment struct {
	Fragment   *Fragment
	Parent     *LFragment
	LocalMeta  *CoreTable
	SharedMeta *CoreTable
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

func (f *LFragment) registerThisFragmentAs(L *lua.LState, name string) {
	ud := L.NewUserData()
	ud.Value = f
	L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
	L.SetGlobal(name, ud)
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
		SharedMeta: NewCoreTable(make(map[string]CoreType)),
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
	"getSharedMeta": fragmentGetSharedMeta,
	"setMeta":       fragmentSetMeta,
	"setSharedMeta": fragmentSetSharedMeta,
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
		L.Push(NewCoreString(f.Fragment.Name).luaType(L))
	case "depth":
		L.Push(NewCoreNumber(float64(f.Fragment.Depth)).luaType(L))
	case "code":
		L.Push(NewCoreString(f.Fragment.Code).luaType(L))
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
	case "sharedMeta":
		if value.Type() == lua.LTTable {
			f.SharedMeta = NewCoreTableL(value.(*lua.LTable))
		} else {
			L.ArgError(3, "table expected for sharedMeta")
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

func fragmentGetSharedMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := getNestedValue(f.SharedMeta, key)
	L.Push(value.luaType(L))
	return 1
}

func fragmentSetMeta(L *lua.LState) int {
	f := checkFragment(L)

	if f.Fragment.EvalState != PENDING {
		log.Warn("Setting metadata on a partially evaluated fragment will not trigger re-evaluation of previous computations.")
	}

	key := L.CheckString(2)
	value := luaToCoreType(L.Get(3))
	setNestedValue(f.LocalMeta, key, value)
	return 0
}

func fragmentSetSharedMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := luaToCoreType(L.Get(3))
	setNestedValue(f.SharedMeta, key, value)
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

const fc = `
-- PLACEHOLDER LUA
=====
Fragment content.

${key}

@{header}

@{sub/test}
`

func testLua() {

	pf := &Fragment{
		Name:       "Parent Fragment",
		Code:       fc,
		Depth:      0,
		Parent:     nil,
		LocalMeta:  *NewEmptyCoreTable(),
		SharedMeta: NewEmptyCoreTable(),
	}

	pf.EvalState = EVALUATING

	pf.LocalMeta.v["key"] = NewCoreString("This is a key")

	pf.RunLua(`print(this:getMeta("key"))`)

}

func (f *Fragment) RunLua(l string) {
	lf := f.MakeLFragment()
	L := lua.NewState()
	defer L.Close()
	registerFragmentType(L)
	libs.Preload(L)
	lf.registerThisFragmentAs(L, "this")

	err := L.DoString(l)
	if err != nil {
		log.Error(err)
	}
}

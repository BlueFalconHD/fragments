package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	lua "github.com/yuin/gopher-lua"
)

type LFragment struct {
	Fragment   *Fragment
	Parent     *LFragment
	LocalMeta  *CoreTable
	SharedMeta *CoreTable
	Builders   *CoreTable
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
		Builders:   NewCoreTable(make(map[string]CoreType)),
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
	L.ArgError(1, fmt.Sprintf("fragment expected, got %s", L.Get(1).Type().String()))
	return nil
}

var fragmentMethods = map[string]lua.LGFunction{
	"getLocalMeta":  fragmentGetMeta,
	"getSharedMeta": fragmentGetSharedMeta,
	"getBothMeta":   fragmentGetBothMeta,
	"setLocalMeta":  fragmentMergeMeta,
	"setSharedMeta": fragmentMergeSharedMeta,
	"parent":        fragmentParent,
	"addBuilders":   fragmentBuilders,
	"builders":      fragmentGetBuilders,
	"setTemplate":   fragmentSetTemplate,
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

func fragmentMergeMeta(L *lua.LState) int {
	// Mergemeta is interesting, it takes in a table and merges it with the current metadata, overwriting existing keys, and creating non-existent ones
	f := checkFragment(L)
	if f.Fragment.EvalState != PENDING {
		log.Warn("Merging metadata on a partially evaluated fragment will not trigger re-evaluation of previous computations.", "fragment", f.Fragment.Name)
	}

	if L.GetTop() < 2 {
		L.ArgError(2, "table expected")
	}

	if L.Get(2).Type() != lua.LTTable {
		L.ArgError(2, "table expected")
	}

	table := L.CheckTable(2)
	gt := NewCoreTableL(table)

	f.LocalMeta.mergeMut(gt)

	return 0
}

func fragmentMergeSharedMeta(L *lua.LState) int {
	f := checkFragment(L)
	if L.GetTop() < 2 {
		L.ArgError(2, "table expected")
	}

	if L.Get(2).Type() != lua.LTTable {
		L.ArgError(2, "table expected")
	}

	table := L.CheckTable(2)
	gt := NewCoreTableL(table)

	f.SharedMeta.mergeMut(gt)

	return 0
}

func fragmentGetBothMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := getNestedValue(f.SharedMeta, key)
	if _, isNil := value.(*CoreNil); isNil {
		value = getNestedValue(f.LocalMeta, key)
	}
	L.Push(value.luaType(L))
	return 1
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

func fragmentBuilders(L *lua.LState) int {
	f := checkFragment(L)
	if L.GetTop() < 2 {
		L.ArgError(2, "table expected")
	}

	if L.Get(2).Type() != lua.LTTable {
		L.ArgError(2, "table expected")
	}

	table := L.CheckTable(2)
	gt := NewCoreTableL(table)

	for key, k := range gt.v {
		lv := k.luaType(L)
		if _, ok := lv.(*lua.LFunction); !ok {
			L.ArgError(2, fmt.Sprintf("expected function at key '%s'", key))
		}
	}

	if f.Fragment.Builders == nil {
		f.Fragment.Builders = NewEmptyCoreTable()
	}
	f.Fragment.Builders.mergeMut(gt)

	return 0
}

func fragmentGetBuilders(L *lua.LState) int {
	f := checkFragment(L)
	if f.Fragment.Builders == nil {
		f.Fragment.Builders = NewEmptyCoreTable()
	}
	L.Push(f.Fragment.Builders.luaType(L))
	return 1
}

func fragmentSetTemplate(L *lua.LState) int {
	f := checkFragment(L)
	if L.GetTop() < 2 {
		L.ArgError(2, "string expected")
	}

	if L.Get(2).Type() != lua.LTString {
		L.ArgError(2, "string expected")
	}

	// Set real fragment's template member to a pointer to the template fragment referenced by name
	t := GetFragmentFromName(L.CheckString(2), TEMPLATE, f.Fragment.FragmentCache)
	f.Fragment.Template = t

	t.FragmentCache = f.Fragment.FragmentCache

	return 0
}

func (f *LFragment) registerThisFragmentAs(L *lua.LState, name string) {
	ud := L.NewUserData()
	ud.Value = f
	L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
	L.SetGlobal(name, ud)
}

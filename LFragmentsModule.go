package main

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
)

type LFragmentsModule struct {
	FragmentCache *FragmentCache
	FragPath      string
	PagePath      string
}

func newFragmentsModule(fragmentCache *FragmentCache, fragPath string, pagePath string) *LFragmentsModule {
	return &LFragmentsModule{
		FragmentCache: fragmentCache,
		FragPath:      fragPath,
		PagePath:      pagePath,
	}
}

const luaFragmentModuleTypeName = "fragments"

func registerFragmentsModuleType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaFragmentModuleTypeName)
	L.SetGlobal("fragments", mt)
	// Methods and metamethods
	L.SetField(mt, "__index", L.NewFunction(fragmentsModuleIndex))
}

func checkFragmentsModule(L *lua.LState) *LFragmentsModule {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*LFragmentsModule); ok {
		return v
	}
	L.ArgError(1, fmt.Sprintf("fragments module expected, got %s", L.Get(1).Type().String()))
	return nil
}

func fragmentsModuleGetFragment(L *lua.LState) int {

	f := checkFragmentsModule(L)
	if L.GetTop() < 2 {
		L.ArgError(2, "string expected")
	}

	if L.Get(2).Type() != lua.LTString {
		L.ArgError(2, "string expected")
	}

	name := L.CheckString(2)

	if f.FragmentCache == nil {
		L.RaiseError("FragmentCache is not initialized.")
		return 0
	}

	fc := f.FragmentCache
	frag := fc.Get(name, FRAGMENT)

	if frag == nil {
		L.RaiseError("Fragment not found: %s", name)
		return 0
	}
	lf := frag.MakeLFragment()

	ud := L.NewUserData()
	ud.Value = lf
	L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
	L.Push(ud)

	return 1
}

func fragmentsModuleGetPage(L *lua.LState) int {
	f := checkFragmentsModule(L)
	if L.GetTop() < 2 {
		L.ArgError(2, "string expected")
	}

	if L.Get(2).Type() != lua.LTString {
		L.ArgError(2, "string expected")
	}

	name := L.CheckString(2)

	if f.FragmentCache == nil {
		L.RaiseError("FragmentCache is not initialized.")
		return 0
	}

	fc := f.FragmentCache
	frag := fc.Get(name, PAGE)

	if frag == nil {
		L.RaiseError("Fragment not found: %s", name)
		return 0
	}
	lf := frag.MakeLFragment()

	ud := L.NewUserData()
	ud.Value = lf
	L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
	L.Push(ud)

	return 1
}

func getFragmentsModuleMethods() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"getFragment": fragmentsModuleGetFragment,
		"getPage":     fragmentsModuleGetPage,
	}
}

func fragmentsModuleIndex(L *lua.LState) int {
	fm := checkFragmentsModule(L)
	field := L.CheckString(2)

	if method, ok := getFragmentsModuleMethods()[field]; ok {
		L.Push(L.NewFunction(method))
		return 1
	}

	switch field {
	case "fragPath":
		L.Push(NewCoreString(fm.FragPath).luaType(L))
	case "pagePath":
		L.Push(NewCoreString(fm.PagePath).luaType(L))
	default:
		L.Push(lua.LNil)
	}

	return 1
}

func fragmentsModuleNewIndex(L *lua.LState) int {
	_ = checkFragmentsModule(L)
	field := L.CheckString(2)
	_ = L.Get(3)

	// do not permit setting of these fields

	switch field {
	case "fragPath":
		L.ArgError(2, "field is read-only")
	case "pagePath":
		L.ArgError(2, "field is read-only")
	default:
		L.ArgError(2, "unexpected field: "+field)
	}
	return 0
}

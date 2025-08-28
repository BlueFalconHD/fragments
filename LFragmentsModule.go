package main

import (
	"fmt"
	"strings"

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
	L.SetField(mt, "__newindex", L.NewFunction(fragmentsModuleNewIndex))
}

func checkFragmentsModule(L *lua.LState) *LFragmentsModule {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*LFragmentsModule); ok {
		return v
	}
	L.ArgError(1, fmt.Sprintf("fragments module expected, got %s", L.Get(1).Type().String()))
	return nil
}

func fragmentsModuleGetAllPages(L *lua.LState) int {
	f := checkFragmentsModule(L)

	// return a table in the format of:
	// {
	//    index = index.frag,
	//    posts = {
	//       example = example.frag,
	//    }
	// }

	if f.FragmentCache == nil {
		L.RaiseError("FragmentCache is not initialized.")
		return 0
	}

	fc := f.FragmentCache
	pages := fc.Cache

	tbl := L.NewTable()
	for name, frag := range pages {
		if frag.Type == PAGE {
			lf := frag.MakeLFragment()
			ud := L.NewUserData()
			ud.Value = lf
			L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
			tbl.RawSetString(name, ud)
		}
	}

	L.Push(tbl)
	return 1

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

func fragmentsModuleGetBuilders(L *lua.LState) int {
	f := checkFragmentsModule(L)
	if L.GetTop() < 3 {
		L.ArgError(2, "expected kind and name")
	}
	kind := L.CheckString(2)
	name := L.CheckString(3)

	if f.FragmentCache == nil {
		L.RaiseError("FragmentCache is not initialized.")
		return 0
	}

	var ft FragmentType
	if kind == "page" {
		ft = PAGE
	} else if kind == "fragment" {
		ft = FRAGMENT
	} else if kind == "template" {
		ft = TEMPLATE
	} else {
		L.ArgError(2, "kind must be 'fragment', 'page', or 'template'")
	}

	frag := f.FragmentCache.Get(name, ft)
	if frag == nil {
		L.RaiseError("Fragment not found: %s", name)
		return 0
	}

	if frag.Builders == nil {
		frag.Builders = NewEmptyCoreTable()
	}
	L.Push(frag.Builders.luaType(L))
	return 1
}

func fragmentsModuleGetPagesUnder(L *lua.LState) int {
	f := checkFragmentsModule(L)
	if L.GetTop() < 2 {
		L.ArgError(2, "string expected")
	}
	if L.Get(2).Type() != lua.LTString {
		L.ArgError(2, "string expected")
	}
	prefix := L.CheckString(2)

	if f.FragmentCache == nil {
		L.RaiseError("FragmentCache is not initialized.")
		return 0
	}

	fc := f.FragmentCache
	tbl := L.NewTable()
	for name, frag := range fc.Cache {
		if frag.Type == PAGE && strings.HasPrefix(name, prefix) {
			lf := frag.MakeLFragment()
			ud := L.NewUserData()
			ud.Value = lf
			L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))

			rel := strings.TrimPrefix(name, prefix)
			if len(rel) > 0 && rel[0] == '/' {
				rel = rel[1:]
			}
			tbl.RawSetString(rel, ud)
		}
	}

	L.Push(tbl)
	return 1
}

func getFragmentsModuleMethods() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"getFragment":   fragmentsModuleGetFragment,
		"getPage":       fragmentsModuleGetPage,
		"getAllPages":   fragmentsModuleGetAllPages,
		"getPagesUnder": fragmentsModuleGetPagesUnder,
		"getBuilders":   fragmentsModuleGetBuilders,
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

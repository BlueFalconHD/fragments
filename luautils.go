package main

// lua utils
// to make using gopher-lua less like being in hell

import (
	lua "github.com/yuin/gopher-lua"
)

func luaToGoType(L *lua.LState, lv lua.LValue) interface{} {
	switch lv.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lv.(lua.LBool)
	case lua.LTNumber:
		// LTNumber is at core a simple float64
		return lv.(lua.LNumber)
	case lua.LTString:
		return lv.(lua.LString)
	case lua.LTTable:
		return luaTableToITTable(L, lv.(*lua.LTable))
	case lua.LTFunction: // this one doesn't really convert that well so we'll just return the ref to the core lua function
		return lv.(*lua.LFunction)
	case lua.LTUserData:
		return lv.(*lua.LUserData).Value
	default:
		// this is a catch-all for types that we don't know how to convert
		return nil
	}
}

func goToLuaType(L *lua.LState, v interface{}) lua.LValue {
	switch v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v.(bool))
	case float64:
		return lua.LNumber(v.(float64))
	case string:
		return lua.LString(v.(string))
	case map[string]interface{}:
		return goMapToITTable(v.(map[string]interface{})).table()
	case *IntermediateTable:
		return v.(*IntermediateTable).table()
	case *lua.LFunction:
		return v.(*lua.LFunction)
	case *lua.LUserData:
		return v.(*lua.LUserData)
	default:
		return lua.LNil
	}
}

// IntermediateTable is a mix between the LTable and map[string]interface{} types
type IntermediateTable struct {
	CoreMap map[string]interface{} // CoreMap stores the actual data
}

func luaTableToITTable(L *lua.LState, lt *lua.LTable) *IntermediateTable {
	coreMap := make(map[string]interface{})

	lt.ForEach(func(k, v lua.LValue) {
		coreMap[k.String()] = luaToGoType(L, v)
	})

	return &IntermediateTable{
		CoreMap: coreMap,
	}
}

func goMapToITTable(m map[string]interface{}) *IntermediateTable {
	return &IntermediateTable{
		CoreMap: m,
	}
}

func (it *IntermediateTable) table() lua.LValue {
	lt := lua.LTable{}

	for k, v := range it.CoreMap {
		lt.RawSetString(k, goToLuaType(nil, v))
	}

	return &lt
}

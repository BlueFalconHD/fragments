package main

import (
	"fmt"
	"github.com/k0kubun/pp/v3"
	lua "github.com/yuin/gopher-lua"
)

type CoreType interface {
	goType() interface{}
	luaType(L *lua.LState) lua.LValue
	stringRepresentation() string
	clone() CoreType
}

type CoreNil struct{}

func NewCoreNil() *CoreNil                          { return &CoreNil{} }
func NewCoreNilL(lv lua.LValue) *CoreNil            { return &CoreNil{} }
func (c *CoreNil) goType() interface{}              { return nil }
func (c *CoreNil) luaType(L *lua.LState) lua.LValue { return lua.LNil }
func (c *CoreNil) stringRepresentation() string     { return "nil" }
func (c *CoreNil) clone() CoreType                  { return NewCoreNil() }

type CoreBool struct{ v bool }

func NewCoreBool(b bool) *CoreBool { return &CoreBool{v: b} }
func NewCoreBoolL(lv lua.LValue) *CoreBool {
	return &CoreBool{v: bool(lv.(lua.LBool))}
}
func (c *CoreBool) goType() interface{}              { return c.v }
func (c *CoreBool) luaType(L *lua.LState) lua.LValue { return lua.LBool(c.v) }
func (c *CoreBool) stringRepresentation() string {
	if c.v {
		return "true"
	}
	return "false"
}
func (c *CoreBool) clone() CoreType { return NewCoreBool(c.v) }

type CoreNumber struct{ v float64 }

func NewCoreNumber(n float64) *CoreNumber { return &CoreNumber{v: n} }
func NewCoreNumberL(lv lua.LValue) *CoreNumber {
	return &CoreNumber{v: float64(lv.(lua.LNumber))}
}
func (c *CoreNumber) goType() interface{}              { return c.v }
func (c *CoreNumber) luaType(L *lua.LState) lua.LValue { return lua.LNumber(c.v) }
func (c *CoreNumber) stringRepresentation() string     { return fmt.Sprintf("%f", c.v) }
func (c *CoreNumber) clone() CoreType                  { return NewCoreNumber(c.v) }

type CoreString struct{ v string }

func NewCoreString(s string) *CoreString { return &CoreString{v: s} }
func NewCoreStringL(lv lua.LValue) *CoreString {
	return &CoreString{v: lv.String()}
}
func (c *CoreString) goType() interface{}              { return c.v }
func (c *CoreString) luaType(L *lua.LState) lua.LValue { return lua.LString(c.v) }
func (c *CoreString) stringRepresentation() string     { return c.v }
func (c *CoreString) clone() CoreType                  { return NewCoreString(c.v) }

type CoreTable struct{ v map[string]CoreType }

func NewEmptyCoreTable() *CoreTable                 { return &CoreTable{v: make(map[string]CoreType)} }
func NewCoreTable(m map[string]CoreType) *CoreTable { return &CoreTable{v: m} }
func NewCoreTableL(lv *lua.LTable) *CoreTable {
	m := make(map[string]CoreType)
	lv.ForEach(func(k, v lua.LValue) {
		m[k.String()] = luaToCoreType(v)
	})
	return &CoreTable{v: m}
}
func (c *CoreTable) goType() interface{} {
	goMap := make(map[string]interface{})
	for k, v := range c.v {
		goMap[k] = v.goType()
	}
	return goMap
}
func (c *CoreTable) luaType(L *lua.LState) lua.LValue {
	lt := L.NewTable()
	for k, v := range c.v {
		lt.RawSetString(k, v.luaType(L))
	}
	return lt
}
func (c *CoreTable) prettyPrint() {
	_, err := pp.Print(c.goType())
	if err != nil {
		fmt.Println(err)
		return
	}
}
func (c *CoreTable) merge(other *CoreTable) {
	for k, v := range other.v {
		if existingVal, exists := c.v[k]; exists {
			existingTable, isExistingTable := existingVal.(*CoreTable)
			otherTable, isOtherTable := v.(*CoreTable)

			if isExistingTable && existingTable != nil && isOtherTable && otherTable != nil {
				// Both are CoreTable and not nil, perform recursive merge
				existingTable.merge(otherTable)
			} else {
				// Either not CoreTable or types differ, overwrite with other.v[k]
				c.v[k] = v
			}
		} else {
			// Key does not exist in c.v, add it
			c.v[k] = v
		}
	}
}
func (c *CoreTable) stringRepresentation() string {
	return fmt.Sprintf("[fragments.CoreTable with %d item%s]", len(c.v), func() string {
		if len(c.v) == 1 {
			return ""
		} else {
			return "s"
		}
	}())
}
func (c *CoreTable) clone() CoreType {
	newMap := make(map[string]CoreType)
	for k, v := range c.v {
		newMap[k] = v.clone()
	}
	return NewCoreTable(newMap)
}

type CoreFunction struct{ v *lua.LFunction }

func NewCoreFunction(f *lua.LFunction) *CoreFunction { return &CoreFunction{v: f} }
func NewCoreFunctionL(lv lua.LValue) *CoreFunction {
	return &CoreFunction{v: lv.(*lua.LFunction)}
}
func (c *CoreFunction) goType() interface{}              { return c.v }
func (c *CoreFunction) luaType(L *lua.LState) lua.LValue { return c.v }
func (c *CoreFunction) stringRepresentation() string     { return "<Lua Function>" }
func (c *CoreFunction) clone() CoreType                  { return NewCoreFunction(c.v) }

type CoreUserData struct{ v *lua.LUserData }

func NewCoreUserData(u *lua.LUserData) *CoreUserData { return &CoreUserData{v: u} }
func (c *CoreUserData) goType() interface{}          { return c.v }
func (c *CoreUserData) luaType(L *lua.LState) lua.LValue {
	return c.v
}
func (c *CoreUserData) stringRepresentation() string { return "<Lua UserData>" }
func (c *CoreUserData) clone() CoreType              { return NewCoreUserData(c.v) }

func luaToCoreType(lv lua.LValue) CoreType {
	switch lv.Type() {
	case lua.LTNil:
		return NewCoreNilL(lv)
	case lua.LTBool:
		return NewCoreBoolL(lv)
	case lua.LTNumber:
		return NewCoreNumberL(lv)
	case lua.LTString:
		return NewCoreStringL(lv)
	case lua.LTTable:
		return NewCoreTableL(lv.(*lua.LTable))
	case lua.LTFunction:
		return NewCoreFunctionL(lv)
	case lua.LTUserData:
		return NewCoreUserData(lv.(*lua.LUserData))
	default:
		return NewCoreNil()
	}
}
func goToCoreType(v interface{}) CoreType {
	switch val := v.(type) {
	case nil:
		return NewCoreNil()
	case bool:
		return NewCoreBool(val)
	case float64:
		return NewCoreNumber(val)
	case int:
		return NewCoreNumber(float64(val))
	case string:
		return NewCoreString(val)
	case map[string]CoreType:
		return NewCoreTable(val)
	case map[string]interface{}:
		coreMap := make(map[string]CoreType)
		for k, v := range val {
			coreMap[k] = goToCoreType(v)
		}
		return NewCoreTable(coreMap)
	case CoreType:
		return val
	default:
		return NewCoreNil()
	}
}

// turn a generic/any type into a string.
func convertToString(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case string:
		return val
	case map[string]CoreType:
		return fmt.Sprintf("%v", val)
	case map[string]interface{}:
		return fmt.Sprintf("%v", val)
	case CoreType:
		return fmt.Sprintf("%v", val)
	default:
		return ""
	}
}

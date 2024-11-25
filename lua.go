package main

import (
	"fmt"
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
	L.ArgError(1, "fragment expected")
	return nil
}

var fragmentMethods = map[string]lua.LGFunction{
	"getMeta":       fragmentGetMeta,
	"getSharedMeta": fragmentGetSharedMeta,
	"setMeta":       fragmentSetMeta,
	"meta":          fragmentMergeMeta,
	"setSharedMeta": fragmentSetSharedMeta,
	"sharedMeta":    fragmentMergeSharedMeta,
	"parent":        fragmentParent,
	"builders":      fragmentBuilders,
	"template":      fragmentSetTemplate,
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

	f.LocalMeta.merge(gt)

	return 0
}

func fragmentSetSharedMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	value := luaToCoreType(L.Get(3))
	setNestedValue(f.SharedMeta, key, value)
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

	f.SharedMeta.merge(gt)

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

func fragmentBuilders(L *lua.LState) int {
	//_ = checkFragment(L)
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

	// Merge the builders table
	// f.Fragment.Builders.merge(gt)
	f.Fragment.Builders = gt

	return 0
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
	t := GetFragmentFromName(L.CheckString(2), TEMPLATE)
	f.Fragment.Template = t

	return 0
}

func customPrint(L *lua.LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {

	}
	fmt.Println()
	return 0
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

// Utility function availible to lua code that renders markdown to html
func renderMarkdown(L *lua.LState) int {
	if L.GetTop() < 1 {
		L.ArgError(1, "string expected")
	}

	if L.Get(1).Type() != lua.LTString {
		L.ArgError(1, "string expected")
	}

	content := L.CheckString(1)
	html, err := RenderMarkdownToHTML(content)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(NewCoreString(html).luaType(L))
	return 1
}

// TODO: remove testing stuff

const fc = `
-- fragments/hello.frag
-- this is the lua body of our fragment

-- we can do some cool stuff here

this:template("post")

function getStringFormattedDate()
    return os.date("%Y-%m-%d")
end

this:meta {
    buildDate = getStringFormattedDate(),
    title = "Hello, World!"
}

this:builders {
    randomBuilder = function()
        -- Pick 40 random characters from the alphabet and append them to a string, then return it
        local alphabet = "abcdefghijklmnopqrstuvwxyz"
        local result = ""
        for i = 1, 10 do
            result = result .. alphabet:sub(math.random(1, #alphabet), math.random(1, #alphabet))
        end
        return result
    end,
	-- Builders can also take content as a single string argument
	reverseBuilder = function(content)
		return string.reverse(content)
	end,
	simpleMarkdownRenderer = function(content)
		-- This is a very simple markdown renderer
		-- It only supports headers and paragraphs
		-- Headers are defined by a line starting with one or more '#' characters
		-- Paragraphs are defined by a line of text
		-- It will return a string with the markdown rendered as HTML
		local result = ""
		local inHeader = false
		for line in content:gmatch("[^\n]+") do
			if line:find("^#+") then
				-- Header
				local headerLevel = line:find("^#+")
				local headerText = line:sub(headerLevel + 1)
				result = result .. "<h" .. headerLevel .. ">" .. headerText .. "</h" .. headerLevel .. ">\n"
			else
				-- Paragraph
				result = result .. "<p>" .. line .. "</p>\n"
			end
		end
		return result
	end
}
---
The content of our fragment begins here.

By using a dollar sign and braces, you can include metadata set in the lua environment: ${buildDate}

Finally, you can dynamically run a lua function that returns a string, like so: *{randomBuilder}

Test reverse builder: *{reverseBuilder[[Hello world!]]}

Test markdown renderer:

*{simpleMarkdownRenderer[[ j

# This is a header

This is a paragraph of text.

# This is another header

This is another paragraph of text.

]]}

Test fragment with content:

@{ihavecontent[[This is the content of the fragment]]}

`

func testLua() {

	pf := &Fragment{
		Name:       "Parent Fragment",
		Code:       fc,
		Depth:      0,
		Parent:     nil,
		LocalMeta:  *NewEmptyCoreTable(),
		SharedMeta: NewEmptyCoreTable(),
		Builders:   NewEmptyCoreTable(),
	}

	pf.LocalMeta.v["key"] = NewCoreString("This is a key")

	log.Info("Output of evaluation", "result", pf.Evaluate())
}

func (f *Fragment) CreateState() *lua.LState {
	lf := f.MakeLFragment()
	L := lua.NewState()
	registerFragmentType(L)

	// Register the markdown rendering function
	L.SetGlobal("renderMarkdown", L.NewFunction(renderMarkdown))

	libs.Preload(L)
	lf.registerThisFragmentAs(L, "this")
	return L
}

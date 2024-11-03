package main

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

type Person struct {
	Name string
	Jobs lua.LValue // Jobs is a Lua table mapping days to functions
}

type LFragment struct {
	// This is the fragment struct built specifically for Lua
	// We have some utility methods here along with pointers to parent LFragments
	// and the actual Fragment struct

	// The actual Fragment struct
	Fragment *Fragment

	// The parent LFragment
	Parent *LFragment

	// Meta data for the fragment
	LocalMeta  IntermediateTable
	GlobalMeta IntermediateTable
}

func (f *LFragment) getFragmentTree() []*LFragment {
	// This method should return a list of LFragment structs from the root (page) to the current LFragment
	return nil
}

const luaPersonTypeName = "person"
const luaFragmentTypeName = "fragment"

// Registers the Person type to the given Lua state.
func registerPersonType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaPersonTypeName)
	L.SetGlobal("person", mt)
	// Static attributes
	L.SetField(mt, "new", L.NewFunction(newPerson))
	// Methods and metamethods
	L.SetField(mt, "__index", L.NewFunction(personIndex))
	L.SetField(mt, "__newindex", L.NewFunction(personNewIndex))
}

func registerFragmentType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaFragmentTypeName)
	L.SetGlobal("fragment", mt)
	// Static attributes
	L.SetField(mt, "new", L.NewFunction(newFragment))
	// Methods and metamethods
	L.SetField(mt, "__index", L.NewFunction(fragmentIndex))
	L.SetField(mt, "__newindex", L.NewFunction(fragmentNewIndex))
}

// Constructor for Person
func newPerson(L *lua.LState) int {
	person := &Person{Name: L.CheckString(1)}
	ud := L.NewUserData()
	ud.Value = person
	L.SetMetatable(ud, L.GetTypeMetatable(luaPersonTypeName))
	L.Push(ud)
	return 1
}

func newFragment(L *lua.LState) int {
	f := &LFragment{
		Fragment:   nil,
		Parent:     nil,
		LocalMeta:  *goMapToITTable(make(map[string]interface{})),
		GlobalMeta: *goMapToITTable(make(map[string]interface{})),
	}
	ud := L.NewUserData()
	ud.Value = f
	L.SetMetatable(ud, L.GetTypeMetatable(luaFragmentTypeName))
	L.Push(ud)
	return 1
}

// Checks whether the first Lua argument is a *LUserData with *Person and returns this *Person.
func checkPerson(L *lua.LState) *Person {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Person); ok {
		return v
	}
	L.ArgError(1, "person expected")
	return nil
}

func checkFragment(L *lua.LState) *LFragment {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*LFragment); ok {
		return v
	}
	L.ArgError(1, "fragment expected")
	return nil
}

var personMethods = map[string]lua.LGFunction{
	"setJobs": personSetJobs,
	"getJob":  personGetJob,
}

var fragmentMethods = map[string]lua.LGFunction{
	"getMeta":       fragmentGetMeta,
	"getGlobalMeta": fragmentGetGlobalMeta,
	"setMeta":       fragmentSetMeta,
	"setGlobalMeta": fragmentSetGlobalMeta,
}

// personIndex handles field access and method calls
func personIndex(L *lua.LState) int {
	p := checkPerson(L)
	field := L.CheckString(2)

	// First, check if field is a method
	if method, ok := personMethods[field]; ok {
		L.Push(L.NewFunction(method))
		return 1
	}

	// Handle properties
	switch field {
	case "name":
		L.Push(lua.LString(p.Name))
	case "jobs":
		if p.Jobs != nil {
			L.Push(p.Jobs)
		} else {
			L.Push(lua.LNil)
		}
	default:
		L.Push(lua.LNil)
	}
	return 1
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
		// throw an error due to the site being a custom type that hasn't been implemented yet
		L.ArgError(2, "site is a custom type that hasn't been implemented yet")
	case "options":
		// ditto
		L.ArgError(2, "options is a custom type that hasn't been implemented yet")
	default:
		L.Push(lua.LNil)
	}
	return 1
}

// personNewIndex handles setting fields
func personNewIndex(L *lua.LState) int {
	p := checkPerson(L)
	field := L.CheckString(2)
	value := L.Get(3)

	switch field {
	case "name":
		p.Name = lua.LVAsString(value)
	case "jobs":
		if value.Type() == lua.LTTable {
			p.Jobs = value
		} else {
			L.ArgError(3, "table expected for jobs")
		}
	default:
		L.ArgError(2, "unexpected field: "+field)
	}
	return 0
}

func fragmentNewIndex(L *lua.LState) int {
	f := checkFragment(L)
	field := L.CheckString(2)
	value := L.Get(3)

	switch field {
	case "localMeta":
		if value.Type() == lua.LTTable {
			// set local meta data
			f.LocalMeta = *luaTableToITTable(L, value.(*lua.LTable))
		} else {
			L.ArgError(3, "table expected for localMeta")
		}
	case "globalMeta":
		if value.Type() == lua.LTTable {
			// set global meta data
			f.GlobalMeta = *luaTableToITTable(L, value.(*lua.LTable))
		} else {
			L.ArgError(3, "table expected for globalMeta")
		}
	default:
		L.ArgError(2, "unexpected field: "+field)
	}
	return 0
}

// Method to set jobs from Lua
func personSetJobs(L *lua.LState) int {
	p := checkPerson(L)
	jobsTable := L.CheckTable(2)
	p.Jobs = jobsTable
	return 0
}

func fragmentGetMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	L.Push(goToLuaType(L, f.LocalMeta.CoreMap[key]))
	return 1
}

func fragmentGetGlobalMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	L.Push(goToLuaType(L, f.GlobalMeta.CoreMap[key]))
	return 1
}

// Method to get a job function by day
func personGetJob(L *lua.LState) int {
	p := checkPerson(L)
	day := L.CheckString(2)
	if p.Jobs == nil {
		L.Push(lua.LNil)
		return 1
	}
	jobsTable := p.Jobs.(*lua.LTable)
	job := jobsTable.RawGetString(day)
	L.Push(job)
	return 1
}

func fragmentSetMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	f.LocalMeta.CoreMap[key] = luaToGoType(L, L.Get(3))
	return 0
}

func fragmentSetGlobalMeta(L *lua.LState) int {
	f := checkFragment(L)
	key := L.CheckString(2)
	//value := L.CheckString(3)
	f.GlobalMeta.CoreMap[key] = luaToGoType(L, L.Get(3))
	return 0
}

func testLua() {
	L := lua.NewState()
	defer L.Close()
	registerPersonType(L)
	registerFragmentType(L)
	if err := L.DoString(`
        p = person.new("Alice")
        print(p.name) -- "Alice"
        p.name = "Bob"
        print(p.name) -- "Bob"

        -- Setting jobs using setJobs method
        p:setJobs {
            monday = function() return "Clean windows" end,
            tuesday = function() return "Write reports" end,
            wednesday = function() return "Attend meetings" end,
            thursday = function() return "Develop features" end,
            friday = function() return "Deploy to production" end,
            saturday = function() return "Maintain servers" end,
            sunday = function() return "Rest day" end,
        }

        -- Accessing jobs table
        print(p.jobs)

        -- Getting a job function and calling it in Lua
        jobFunc = p:getJob("monday")
        print(jobFunc()) -- Should print "Clean windows"

		-- Creating a fragment
		f = fragment.new()
		f:setMeta("title", "Hello, world!")
		print(f:getMeta("title")) -- "Hello, world!"
		f:setGlobalMeta("isCool", true)

    `); err != nil {
		panic(err)
	}

	// Retrieve the Person object from Lua
	ud := L.GetGlobal("p")
	luaud, ok := ud.(*lua.LUserData)
	if !ok {
		fmt.Println("p is not a userdata")
		return
	}
	person, ok := luaud.Value.(*Person)
	if !ok {
		fmt.Println("ud.Value is not a *Person")
		return
	}

	// Access and call a job function for a given day from Go
	if person.Jobs != nil {
		jobsTable := person.Jobs.(*lua.LTable)
		day := "monday"
		jobFunc := jobsTable.RawGetString(day)
		if jobFunc.Type() == lua.LTFunction {
			L.Push(jobFunc)
			err := L.PCall(0, 1, nil)
			if err != nil {
				fmt.Println("Error calling job function:", err)
				return
			}
			ret := L.Get(-1)
			fmt.Printf("Job on %s: %s\n", day, ret.String())
			L.Pop(1)
		} else {
			fmt.Printf("No job function for %s\n", day)
		}
	} else {
		fmt.Println("No jobs set")
	}

	// Retrieve the Fragment object from Lua
	ud = L.GetGlobal("f")
	luaud, ok = ud.(*lua.LUserData)
	if !ok {
		fmt.Println("f is not a userdata")
		return
	}

	fragment, ok := luaud.Value.(*LFragment)
	if !ok {
		fmt.Println("ud.Value is not a *LFragment")
		return
	}

	// Access the meta data for the fragment
	fmt.Println("Fragment title:", fragment.LocalMeta.CoreMap["title"])
	// Access the global meta data for the fragment
	fmt.Println("Fragment isCool:", fragment.GlobalMeta.CoreMap["isCool"])

	// Set the meta data for the fragment
	fragment.LocalMeta.CoreMap["title"] = "Hello, Lua!"
	// Set the global meta data for the fragment
	fragment.GlobalMeta.CoreMap["isCool"] = false
	// Set a table in the meta data
	fragment.LocalMeta.CoreMap["table"] = map[string]interface{}{
		"key": "value",
	}

	// Run a Lua script that prints the fragment's title and isCool status
	if err := L.DoString(`
		print(f:getMeta("title"))
		print(f:getGlobalMeta("isCool"))
		-- Access the table in the meta data
		print(f:getMeta("table").key)
		f:setMeta("table", {key = "new value"})
	`); err != nil {
		panic(err)
	}

	// Access the value of the table in the meta data
	fmt.Println("Fragment table key:", fragment.LocalMeta.CoreMap["table"].(*IntermediateTable).CoreMap["key"])
}

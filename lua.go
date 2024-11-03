package main

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
)

// TODO: 	luar might not be the best choice, might just have to do it manually using gopher-lua's state API
//			cannot access the array of friends from Lua for some reason, it is nil

type User struct {
	Name  string
	token string

	// pointers to friends
	friends []*User
}

func (u *User) SetToken(t string) {
	u.token = t
}

func (u *User) Token() string {
	return u.token
}

const script = `
print("Hello from Lua, " .. tim.Name .. "!")
tim:SetToken("12345")

-- Get tim's friend's name
print("Tim's friend is " .. tim.friends)
`

func testLua() {
	L := lua.NewState()
	defer L.Close()

	// make 2 users, Tim and Joe
	// they are friends

	tim := &User{
		Name:    "Tim",
		friends: []*User{},
	}

	joe := &User{
		Name:    "Joe",
		friends: []*User{},
	}

	tim.friends = append(tim.friends, joe)
	joe.friends = append(joe.friends, tim)

	L.SetGlobal("tim", luar.New(L, tim))
	L.SetGlobal("joe", luar.New(L, joe))
	if err := L.DoString(script); err != nil {
		panic(err)
	}

	fmt.Println("Lua set your token to:", tim.Token())
	// Output:
	// Hello from Lua, Tim!
	// Lua set your token to: 12345
}

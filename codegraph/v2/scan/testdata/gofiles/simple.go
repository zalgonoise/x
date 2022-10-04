package main

import (
	"fmt"
)

type Person interface {
	Say(string)
	Do(string)
}

type Admin interface {
	Person
	Admin(func(*Person))
}

type Human struct {
	Age  int
	Name string
}

func (h *Human) Say(s string) {
	fmt.Println(h.Name, "said", s)
}

func (h *Human) Do(s string) {
	fmt.Println(h.Name, "is doing", s)
}

type Administrator struct {
	Human
	Role string
}

func (a *Administrator) Say(s string) {
	fmt.Println(a.Name, "said", s)
}

func (a *Administrator) Do(s string) {
	fmt.Println(a.Name, "is doing", s)
}

func (a *Administrator) Admin(fn func()) {
	if a.Role == "admin" || a.Role == "manager" {
		fn()
	}
}

func main() {
	h := &Human{
		Age:  26,
		Name: "Go",
	}

	a := &Administrator{
		Human: Human{
			Age:  31,
			Name: "Gopher",
		},
		Role: "admin",
	}

	h.Do("something")

	a.Say("hold it, name change")
	a.Admin(func() {
		h.Name = "Grizzly Fur"
	})
	a.Say("voila!")
	h.Say("damn son where did you find this?")
}

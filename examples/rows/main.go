package main

import (
	"fmt"
	"github.com/medivhyang/bear"
)

func main() {
	//fmt.Println("demo map slice:")
	//demoMapSlice()
	//
	//fmt.Println("\ndemo map:")
	//demoMap()
	//
	//fmt.Println("\ndemo struct slice:")
	//demoStructSlice()
	//
	//fmt.Println("\ndemo struct:")
	//demoStruct()

	demoPaging()
}

func demoMapSlice() {
	rows, err := bear.SelectWithStruct(user{}).Query(getDB())
	if err != nil {
		panic(err)
	}
	slice, err := rows.MapSlice()
	if err != nil {
		panic(err)
	}
	for i, v := range slice {
		fmt.Printf("%d => %#v\n", i, v)
	}
}

func demoMap() {
	rows, err := bear.SelectWithStruct(user{}).Query(getDB())
	if err != nil {
		panic(err)
	}
	m, err := rows.Map()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", m)
}

func demoStructSlice() {
	rows, err := bear.SelectWithStruct(user{}).Query(getDB())
	if err != nil {
		panic(err)
	}
	var users []user
	if err := rows.StructSlice(&users); err != nil {
		panic(err)
	}
	for i, v := range users {
		fmt.Printf("%d => %#v\n", i, v)
	}
}

func demoStruct() {
	rows, err := bear.SelectWithStruct(user{}).Query(getDB())
	if err != nil {
		panic(err)
	}
	var u user
	if err := rows.Struct(&u); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", u)
}

func demoPaging() {
	rows, err := bear.SelectWithStruct(user{}).Query(getDB())
	if err != nil {
		panic(err)
	}
	slice, err := rows.MapSlice()
	if err != nil {
		panic(err)
	}
	for i, v := range slice {
		fmt.Printf("%d => %#v\n", i, v)
	}
}

package main

import (
	"fmt"
	"github.com/medivhyang/bear"
)

func main() {
	b := bear.NewDDLBuilder().CreateTable(bear.Table{
		Name: "user",
		Columns: []bear.Column{
			{Name: "id", Type: "varchar(64)"},
			{Name: "age", Type: "integer"},
		},
	}, true)
	t := b.Pretty("", "  ").Build()
	fmt.Println(t)
}

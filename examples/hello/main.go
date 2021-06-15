package main

import (
	"fmt"
	"github.com/medivhyang/bear"

	_ "github.com/medivhyang/bear/dialect/sqlite3"
)

func main() {
	s := bear.NewBuilder().Select("user", "id", "name", "age").Build()
	fmt.Println(s)
}

package main

import (
	"fmt"
	"github.com/medivhyang/bear"
	"github.com/medivhyang/bear/expr"
)

func main() {
	e := expr.Empty()
	e = e.And(expr.Equal("role", expr.Value("developer")))
	e = e.And(expr.Between("age", expr.Value(20), expr.Value(30)))
	s := bear.Select("user", "id", "age", "role").Where(e.Tuple()).Build()
	fmt.Printf("%+v", s)
}

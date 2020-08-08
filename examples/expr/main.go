package main

import (
	"fmt"
	"github.com/medivhyang/bear"
	"github.com/medivhyang/bear/expr"
)

func main() {
	r := bear.Select("user", "id", "age", "role").
		Where(expr.Equal("role", expr.Value("developer")).Tuple()).
		WhereWithTuple(expr.Between("age", expr.Value(20), expr.Value(30)).Tuple()).
		Build()
	fmt.Printf("%+v", r)
}

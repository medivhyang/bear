package expr

import (
	"fmt"
	"github.com/medivhyang/bear"
)

func ExampleWhereWithTemplate() {
	t := bear.Select("user", "id", "name", "age").WhereWithTemplate(GreaterEqual("age", 20)).Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// select id,name,age from user where (age >= ?)
	// [20]
}

func ExampleHavingWithTemplate() {
	t := bear.Select("user", "role", "count(*) as count").
		GroupBy("role").
		HavingWithTemplate(GreaterEqual("age", 20)).
		Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// select role,count(*) as count from user group by role having (age >= ?)
	// [20]
}

package expr

import (
	"fmt"
	"github.com/medivhyang/bear"
)

func ExampleExprEqual() {
	t := bear.Select("user", "id", "name", "age").
		WhereTemplate(GreaterEqual("id", 1)).
		Build()
	fmt.Println(t)

	// Output:
	// "select id,name,age from user where (id >= ?)" <= {1}
}

func ExampleExprIn() {
	t := bear.Select("user", "id", "name", "age").
		WhereTemplate(In("id", []int{1, 2, 3})).
		Build()
	fmt.Println(t)

	// Output:
	// "select id,name,age from user where (id in (?,?,?))" <= {1, 2, 3}
}

func ExampleExprInSubQuery() {
	t := bear.Select("user", "id", "name", "age").
		WhereTemplate(In("id",
			bear.Select("group", "user_id").
				WhereTemplate(Equal("id", 1)).
				Build(),
		)).
		Build()
	fmt.Println(t)

	// Output:
	// "select id,name,age from user where (id in (select user_id from group where (id = ?)))" <= {1}
}

func ExampleExprBetween() {
	t := bear.Select("user", "id", "name", "age").
		WhereTemplate(Between("age", 20, 30)).
		Build()
	fmt.Println(t)

	// Output:
	// "select id,name,age from user where (age between ? and ?)" <= {20, 30}
}

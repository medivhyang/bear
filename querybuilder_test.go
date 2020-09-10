package bear

import "fmt"

func ExampleSelect() {
	t := Select("user", "id", "name", "age").Where("age > ?", 20).Build()
	fmt.Println(t)

	// Output:
	// "select id,name,age from user where (age > ?)" <= {20}
}

func ExampleSelectStruct() {
	type user struct {
		ID   int
		Name string
		Age  string
	}
	t := SelectStruct(user{}).Where("age > ?", 20).Build()
	fmt.Println(t)

	// Output:
	// "select id,name,age from user where (age > ?)" <= {20}
}

func ExampleSelectWhere() {
	type user struct {
		ID   int
		Name string
		Age  int
	}
	t := SelectWhere(user{ID: 1}).Build()
	fmt.Println(t)

	// Output:
	// "select id,name,age from user where (id = ?)" <= {1}
}

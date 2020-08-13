package bear

import "fmt"

func ExampleSelect() {
	t := Select("user", "id", "name", "age").Where("age > ?", 20).Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// select id,name,age from user where (age > ?)
	// [20]
}

func ExampleSelectWithStruct() {
	type user struct {
		ID   int
		Name string
		Age  string
	}
	t := SelectWithStruct(user{}).Where("age > ?", 20).Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// select id,name,age from user where (age > ?)
	// [20]
}

func ExampleSelectWhere() {
	type user struct {
		ID   int
		Name string
		Age  int
	}
	t := SelectWhere(user{ID: 1}).Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// select id,name,age from user where (id = ?)
	// [1]
}

func ExampleInsert() {
	t := Insert("user", map[string]interface{}{"id": 1}).Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// insert into user(id) values(?)
	// [1]
}

func ExampleUpdate() {
	t := Update("user", map[string]interface{}{"name": "new_name"}).Where("id = ?", 1).Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// update user set name=? where (id = ?)
	// [new_name 1]
}

func ExampleDelete() {
	t := Delete("user").Where("id = ?", 1).Build()
	fmt.Println(t.Format)
	fmt.Println(t.Values)

	// Output:
	// delete from user where (id = ?)
	// [1]
}

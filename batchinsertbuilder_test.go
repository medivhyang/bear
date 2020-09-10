package bear

import "fmt"

func ExampleBatchInsert() {
	data := []map[string]interface{}{
		{"id": 1, "name": "alice"},
		{"id": 2, "name": "bob"},
		{"id": 3, "name": "lisa"},
		{"id": 4},
		{"name": "Kitty"},
		{},
	}
	t := BatchInsert("user", data).Build()
	fmt.Println(t)

	// Output:
	// "insert into user(id,name) values(?,?)(?,?)(?,?)(?,?)(?,?)(?,?)" <= {1, "alice", 2, "bob", 3, "lisa", 4, <nil>, <nil>, "Kitty", <nil>, <nil>}
}

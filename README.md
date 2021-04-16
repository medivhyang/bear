# Bear

A go language sql builder.

## Why Bear

Lightweight、lower dependence and efficient.

## Quick Start

```go
package main

import (
	"fmt"
	"time"

	"github.com/medivhyang/bear"
)

type user struct {
	ID      int    `bear:"name=id,type=integer,suffix=primary key"`
	Name    string `bear:"name=name,type=text,suffix=not null"`
	Created int64  `bear:"type=integer,suffix=not null"`
}

func ExampleSelect() {
	t := bear.Select("user", "id", "name").Where("id = ?", 1).Build()
	fmt.Println(t)

	// Output:
	// "select id,name from user where (id = ?)" <= {1}
}

func ExampleSelectStruct() {
	t := bear.Select("user", user{}).Where("id = ?", 1).Build()
	fmt.Println(t)

	// Output:
	// "select id,name,created from user where (id = ?)" <= {1}
}

func ExampleInsert() {
	t := bear.Insert("user", map[string]interface{}{
		"id":      1,
		"name":    "alice",
		"created": time.Date(2020, 9, 11, 0, 0, 0, 0, time.UTC).Unix(),
	}).Build()
	fmt.Println(t)

	// Output:
	// "insert into user(id,name,created) values(?,?,?)" <= {1, "alice", 1599782400}
}

func ExampleInsertStruct() {
	t := bear.Insert("user", user{
		ID:      1,
		Name:    "bob",
		Created: time.Date(2020, 9, 11, 0, 0, 0, 0, time.UTC).Unix(),
	}, true).Build()
	fmt.Println(t)

	// Output:
	// "insert into user(id,name,created) values(?,?,?)" <= {1, "alice", 1599782400}
}

func ExampleUpdate() {
	t := bear.Update("user", map[string]interface{}{"name": "new alice"}).
		Where("id = ?", 1).
		Build()
	fmt.Println(t)

	// Output:
	// "update user set name=? where (id = ?)" <= {"new alice", 1}
}

func ExampleUpdateStruct() {
	t := bear.Update("user", user{Name: "new alice"}, true).
		Where("id = ?", 1).
		Build()
	fmt.Println(t)

	// Output:
	// "update user set name=? where (id = ?)" <= {"new alice", 1}
}

func ExampleDelete() {
	t := bear.Delete("user").Where("id = ?", 1).Build()
	fmt.Println(t)

	// Output:
	// "delete from user where (id = ?)" <= {1}
}

func ExampleDelete() {
	t := bear.Delete("user").Where("id = ?", 1).Build()
	fmt.Println(t)

	// Output:
	// "delete from user where (id = ?)" <= {1}
}
```

> More examples refer to `/examples`

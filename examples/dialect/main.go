package main

import (
	"fmt"
	"time"

	"github.com/medivhyang/bear"
	_ "github.com/medivhyang/bear/dialect/sqlite3"
)

func main() {
	type user struct {
		Name        string
		Age         int
		Created     time.Time
		IgnoreField string `bear:"-"`
	}

	fmt.Printf(bear.CreateTableWithStructIfNotExists(user{}).Dialect("sqlite3").Build().Format)

	// or
	// bear.SetDefaultDialect("sqlite3")
	// fmt.Printf(bear.CreateTableWithStructIfNotExists(user{}).Build().Format)
}

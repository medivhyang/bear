package main

import (
	"fmt"
	"github.com/medivhyang/bear"
)

func main() {
	b := bear.NewBulkInsertBuilder().Table("user").Columns("id", "age")
	b.AppendMap(map[string]interface{}{
		"id":  "Medivh",
		"age": 24,
	})
	b.Append([]interface{}{"Jason", 27})
	fmt.Println(b.Build())
}

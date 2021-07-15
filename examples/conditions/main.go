package main

import (
	"fmt"

	"github.com/medivhyang/bear"
	_ "github.com/medivhyang/bear/dialect/sqlite3"
)

func main() {
	cc := bear.NewConditions()
	cc = cc.Appendf("name = ?", "Medivh")
	cc = cc.Appendf("age = ?", 20)
	cc = cc.AppendMap(map[string]interface{}{
		"human": true,
	})
	cc = cc.AppendIn("likes", "cat", "dog", "tiger")
	fmt.Println(cc.JoinAnd())
}

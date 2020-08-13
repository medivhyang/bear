package main

import (
	"fmt"
	"time"

	"github.com/medivhyang/bear"
	_ "github.com/medivhyang/bear/dialect/sqlite3"
	"github.com/medivhyang/bear/expr"
)

func main() {
	items := []struct {
		prefix   string
		template bear.Template
	}{
		{prefix: "demo select:", template: demoSelect()},
		{prefix: "demo select with struct:", template: demoSelectWithStruct()},
		{prefix: "demo select where:", template: demoSelectWhere()},
		{prefix: "demo select join:", template: demoSelectJoin()},
		{prefix: "demo select sub query:", template: demoSelectSubQuery()},
		{prefix: "demo insert:", template: demoInsert()},
		{prefix: "demo update:", template: demoUpdate()},
		{prefix: "demo delete:", template: demoDelete()},
		{prefix: "demo dialect:", template: demoDialect()},
	}

	for index, item := range items {
		if index > 0 {
			fmt.Println()
		}
		fmt.Println(item.prefix)
		fmt.Printf("%#v\n", item.template)
	}
}

type order struct {
	ID     int `bear:"type=integer"`
	UserID int `bear:"type=integer"`
}

type user struct {
	ID          int    `bear:"type=integer"`
	Name        string `bear:"type=text"`
	Age         int    `bear:"type=integer"`
	Role        string `bear:"type=text"`
	Created     int64  `bear:"type=integer"`
	IgnoreField string `bear:"-"`
}

func demoSelect() bear.Template {
	return bear.Select("user", "id", "name", "age", "role", "created").Where("id = ?", 1).Build()
}

func demoSelectWithStruct() bear.Template {
	return bear.SelectWithStruct(user{}).Where("id = ?", 1).Build()
}

func demoSelectWhere() bear.Template {
	return bear.SelectWhere(user{ID: 1}).Build()
}

func demoSelectJoin() bear.Template {
	return bear.Select(bear.TableName(user{}), "order.id", "order.user_id", "user.name").
		Join("left join order on user.id = order.user_id").
		Where("user.name = ?", "Alice").
		Build()
}

func demoSelectSubQuery() bear.Template {
	return bear.SelectWithStruct(user{}).
		WhereWithTemplate(expr.GreaterEqualTemplate("age", bear.Select("user", "avg(age)").Build())).
		Build()
}

func demoInsert() bear.Template {
	return bear.InsertWithStruct(user{
		ID:      1,
		Name:    "Medivh",
		Age:     20,
		Role:    "teacher",
		Created: time.Now().Unix(),
	}).Build()
}

func demoUpdate() bear.Template {
	return bear.UpdateWithStruct(user{Name: "New Name"}).Where("id = ?", 1).Build()
}

func demoDelete() bear.Template {
	return bear.Delete("user").Where("id = ?", 1).Build()
}

func demoDialect() bear.Template {
	type foo struct {
		Name        string
		Age         int
		Created     time.Time
		IgnoreField string `bear:"-"`
	}

	return bear.CreateTableWithStructIfNotExists(foo{}).Dialect("sqlite3").Build()

	// or
	// bear.SetDefaultDialect("sqlite3")
	// return bear.CreateTableWithStructIfNotExists(foo{}).Build()
}
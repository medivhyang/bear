package main

import (
	"fmt"
	"github.com/medivhyang/bear"
	"github.com/medivhyang/bear/expr"
)

func main() {
	demoSelect()
	demoSelectJoin()
	demoSelectSubQuery()
}

type order struct {
	ID     int `bear:"type=integer"`
	UserID int `bear:"type=integer"`
}

type user struct {
	ID      int    `bear:"name=id,type=integer"`
	Name    string `bear:"name=name,type=text"`
	Age     int    `bear:"name=age,type=integer"`
	Role    string `bear:"name=role,type=text"`
	Created int64  `bear:"name=created,type=integer"`
	Ignore  string `bear:"-"`
}

func demoSelect() {
	t := bear.SelectWithStruct(user{}).
		WhereWithTemplate(expr.Equal("name", "Alice")).
		WhereWithTemplate(expr.GreaterThan("age", 20)).
		Build()
	fmt.Printf("%#v\n", t)
}

func demoSelectJoin() {
	t := bear.SelectSimple(bear.TableName(user{}), "order.id", "order.user_id", "user.name").
		Join("left join order on user.id = order.user_id").
		WhereWithTemplate(expr.Equal("user.name", "Alice")).
		Build()
	fmt.Printf("%#v\n", t)
}

func demoSelectSubQuery() {
	t := bear.SelectWithStruct(user{}).
		WhereWithTemplate(expr.GreaterEqualTemplate("age", bear.SelectSimple(bear.TableName(user{}), "avg(age)").Build())).
		Build()
	fmt.Printf("%#v\n", t)
}

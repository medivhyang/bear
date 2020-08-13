# Bear

Bear is a sql builder.

## Why Bear

1. Lightweight, not much new concept.
2. Native, compatible with standard libary.
3. Rich features, support dialect, struct binding and complex query etc.
4. Efficient, follow engineering practice.
5. Templating, the core concept of Bear is template.

## Quick Start

Common type define

```
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
}
```

Select:

```go
bear.Select("user", "id", "name", "age", "role", "created").Where("id = ?", 1).Build()

// sql: select id,name,age,role,created from user where (id = ?)
// values: {1}
```

Select with struct:

```go
bear.SelectWithStruct(user{}).Where("id = ?", 1).Build()

// sql: select id,name,age,role,created from user where (id = ?)
// values: {1}
```

Select where

```go
bear.SelectWhere(user{ID: 1}).Build()

// sql: select id,name,age,role,created from user where (id = ?)
// values: {1}
```

Select join:

```go
bear.Select(bear.TableName(user{}), "order.id", "order.user_id", "user.name").
		Join("left join order on user.id = order.user_id").
		Where("user.name = ?", "Alice").
		Build()

// sql: select order.id,order.user_id,user.name from user left join order on user.id = order.user_id where (user.name = ?)
// values: {"Alice"}
```

Select sub query:

```go
bear.SelectWithStruct(user{}).
		WhereWithTemplate(expr.GreaterEqualTemplate("age", bear.Select("user", "avg(age)").Build())).
		Build()

// sql: select id,name,age,role,created from user where (age >= (select avg(age) from user))
```

Insert:

```go
bear.InsertWithStruct(user{
		ID:      1,
		Name:    "Alice",
		Age:     25,
		Role:    "teacher",
		Created: time.Now().Unix(),
	}).Build()
// sql: insert into user(role,created,id,name,age) values(?,?,?,?,?)
// values: {"teacher", 1597288323, 1, "Medivh", 20}
```

Update:

```go
bear.UpdateWithStruct(user{Name: "New Name"}).Where("id = ?", 1).Build()

// result: 
// sql: update user set name=? where (id = ?)
// values: {"New Name", 1}}
```

Delete:

```go
bear.Delete("user").Where("id = ?", 1).Build()

// result:
// sql: delete from user where (id = ?)
// values: {1}
```

Dialect:

```go
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

// result:
// sql:
// create table if not exists foo (
//  name text,
//  age integer,
//  created datetime
// );
```
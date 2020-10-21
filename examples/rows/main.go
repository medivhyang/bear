package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/medivhyang/bear"
	_ "github.com/medivhyang/bear/dialect/sqlite3"
)

type user struct {
	ID      int    `bear:"suffix:primary key"`
	Name    string `bear:"suffix:not null"`
	Age     int    `bear:"suffix:not null"`
	Role    string `bear:"suffix:not null"`
	Created int64  `bear:"column:create_time,suffix:not null"`
}

var db *sql.DB

func init() {
	bear.EnableDebug(false)

	var err error
	db, err = openSqlite3("data.db")
	if err != nil {
		panic(err)
	}
	if _, err := bear.DropTable("user").
		OnExists().
		Execute(db); err != nil {
		panic(err)
	}
	if _, err := bear.CreateTableStruct("user", user{}).
		OnNotExists().
		Execute(db); err != nil {
		panic(err)
	}
	data := []user{
		{ID: 1, Name: "Tom", Age: 20, Role: "student", Created: time.Now().Unix()},
		{ID: 2, Name: "Bob", Age: 21, Role: "student", Created: time.Now().Unix()},
		{ID: 3, Name: "Medivh", Age: 32, Role: "teacher", Created: time.Now().Unix()},
		{ID: 4, Name: "Jason", Age: 33, Role: "teacher", Created: time.Now().Unix()},
		{ID: 5, Name: "Monica", Age: 34, Role: "teacher", Created: time.Now().Unix()},
	}
	if _, err := bear.BatchInsertStruct(data).Execute(db); err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("demo map slice:")
	demoMapSlice(db)

	fmt.Println("\ndemo map:")
	demoMap(db)

	fmt.Println("\ndemo struct slice:")
	demoStructSlice(db)

	fmt.Println("\ndemo struct:")
	demoStruct(db)
}

func openSqlite3(filename string) (*sql.DB, error) {
	if _, err := os.Stat(filename); !os.IsExist(err) {
		if _, err := os.Create(filename); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func demoMapSlice(db *sql.DB) {
	rows, err := bear.SelectStruct("user", user{}).Query(db)
	if err != nil {
		panic(err)
	}
	slice, err := rows.MapSlice()
	if err != nil {
		panic(err)
	}
	for i, v := range slice {
		fmt.Printf("%d => %#v\n", i, v)
	}
}

func demoMap(db *sql.DB) {
	rows, err := bear.SelectStruct("user", user{}).Query(db)
	if err != nil {
		panic(err)
	}
	m, err := rows.Map()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", m)
}

func demoStructSlice(db *sql.DB) {
	rows, err := bear.SelectStruct("user", user{}).Query(db)
	if err != nil {
		panic(err)
	}
	var users []user
	if err := rows.StructSlice(&users); err != nil {
		panic(err)
	}
	for i, v := range users {
		fmt.Printf("%d => %#v\n", i, v)
	}
}

func demoStruct(db *sql.DB) {
	rows, err := bear.SelectStruct("user", user{}).Query(db)
	if err != nil {
		panic(err)
	}
	var u user
	if err := rows.Struct(&u); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", u)
}

package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"time"
)

const (
	DBPath      = "data.db"
	initSQLPath = "init.sql"
)

type User struct {
	ID      int
	Name    string
	Age     int
	Role    string
	Created int64
}

func main() {
	if _, err := os.Stat(DBPath); !os.IsExist(err) {
		os.Create(DBPath)
	}
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	bs, err := ioutil.ReadFile(initSQLPath)
	if err != nil {
		panic(err)
	}
	if _, err := db.Exec(string(bs)); err != nil {
		panic(err)
	}
	data := initData()
	for _, item := range data {
		_, err := db.Exec("insert into user(name, age, role, created) values(?, ?, ?, ?)",
			item.Name, item.Age, item.Role, item.Created)
		if err != nil {
			panic(err)
		}
	}
}

func initData() []User {
	return []User{
		{ID: 1, Name: "Tom", Age: 20, Role: "student", Created: time.Now().Unix()},
		{ID: 2, Name: "Bob", Age: 21, Role: "student", Created: time.Now().Unix()},
		{ID: 3, Name: "Medivh", Age: 32, Role: "teacher", Created: time.Now().Unix()},
		{ID: 4, Name: "Jason", Age: 33, Role: "teacher", Created: time.Now().Unix()},
		{ID: 5, Name: "Monica", Age: 34, Role: "teacher", Created: time.Now().Unix()},
	}
}

package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	_dbPath  = "data.db"
	_initSQL = `
drop table if exists user;
create table if not exists user (
    id integer primary key,
    name text,
    age integer,
    role varchar(64),
    created bigint
);
`
)

var (
	_data = []user{
		{ID: 1, Name: "Tom", Age: 20, Role: "student", Created: time.Now().Unix()},
		{ID: 2, Name: "Bob", Age: 21, Role: "student", Created: time.Now().Unix()},
		{ID: 3, Name: "Medivh", Age: 32, Role: "teacher", Created: time.Now().Unix()},
		{ID: 4, Name: "Jason", Age: 33, Role: "teacher", Created: time.Now().Unix()},
		{ID: 5, Name: "Monica", Age: 34, Role: "teacher", Created: time.Now().Unix()},
	}
	_dbIns *sql.DB
)

type user struct {
	ID      int    `json:"id" bear:"name=id"`
	Name    string `json:"name" bear:"name=name"`
	Age     int    `json:"age" bear:"name=age"`
	Role    string `json:"role" bear:"name=role"`
	Created int64  `json:"created" bear:"name=created"`
}

func init() {
	if err := initDB(); err != nil {
		panic(err)
	}
}

func initDB() error {
	if _, err := os.Stat(_dbPath); !os.IsExist(err) {
		if _, err := os.Create(_dbPath); err != nil {
			return err
		}
	}
	db, err := sql.Open("sqlite3", _dbPath)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	if _, err := db.Exec(_initSQL); err != nil {
		return err
	}
	for _, item := range _data {
		_, err := db.Exec("insert into user(name, age, role, created) values(?, ?, ?, ?)",
			item.Name, item.Age, item.Role, item.Created)
		if err != nil {
			return err
		}
	}

	_dbIns = db

	return nil
}

func getDB() *sql.DB {
	return _dbIns
}

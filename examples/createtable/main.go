package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/medivhyang/bear"
)

type user struct {
	ID      int    `bear:"name=id,type=integer"`
	Name    string `bear:"name=name,type=text"`
	Age     int    `bear:"name=age,type=integer"`
	Role    string `bear:"name=role,type=text"`
	Created int64  `bear:"name=created,type=integer"`
}

func main() {
	db, err := openSqliteDB("data.db")
	if err != nil {
		panic(err)
	}
	if _, err := bear.DropTable("user").Execute(db); err != nil {
		panic(err)
	}
	if _, err := bear.CreateTableWithStruct(user{}).Execute(db); err != nil {
		panic(err)
	}
	fmt.Println("done")
}

func openSqliteDB(filename string) (*sql.DB, error) {
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

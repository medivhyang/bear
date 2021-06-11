package bear

import (
	"context"
	"database/sql"
)

//type Raw interface {
//	Query(query string, args ...interface{}) (*sql.rows, error)
//	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.rows, error)
//	Exec(query string, args ...interface{}) (sql.Result, error)
//	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
//}

type DB struct {
	Raw *sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{Raw: db}
}

func OpenDB(driverName string, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return NewDB(db), err
}

func (db *DB) Query(ctx context.Context, t Template, i interface{}) error {
	debugf("bear: query: %s", t.String())
	rows, err := db.Raw.QueryContext(ctx, t.Format, t.Values...)
	if err != nil {
		return err
	}
	return NewRows(rows).Bind(i)
}

func (db *DB) Exec(ctx context.Context, t Template) (sql.Result, error) {
	debugf("bear: exec: %s", t.String())
	return db.Raw.ExecContext(ctx, t.Format, t.Values...)
}

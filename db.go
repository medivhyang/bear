package bear

import (
	"context"
	"database/sql"
)

type Raw interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type DB interface {
	Query(ctx context.Context, t Template, i interface{}) error
	Exec(ctx context.Context, t Template) (sql.Result, error)
	Tx(ctx context.Context, fn func(ctx context.Context, tx DB) error) (err error)
	BeginTx(ctx context.Context) (DB, error)
	Rollback() error
	Commit() error
}

type db struct {
	raw Raw
}

func NewDB(r Raw) DB {
	return &db{raw: r}
}

func OpenDB(driverName string, dataSourceName string) (DB, error) {
	r, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := r.Ping(); err != nil {
		return nil, err
	}
	return NewDB(r), err
}

func (db *db) Query(ctx context.Context, t Template, i interface{}) error {
	debugf("query: %s", t.String())
	rows, err := db.raw.QueryContext(ctx, t.Format, t.Values...)
	if err != nil {
		return err
	}
	return NewRows(rows).Bind(i)
}

func (db *db) Exec(ctx context.Context, t Template) (sql.Result, error) {
	debugf("exec: %s", t.String())
	return db.raw.ExecContext(ctx, t.Format, t.Values...)
}

func (db *db) Tx(ctx context.Context, fn func(ctx context.Context, tx DB) error) (err error) {
	var tx DB
	if tx, err = db.BeginTx(ctx); err != nil {
		return err
	}
	defer func() {
		if x := recover(); x != nil {
			debugf("tx", "catch panic: %v", x)
			err = tx.Rollback()
		}
	}()
	if err := fn(ctx, tx); err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

func (db *db) BeginTx(ctx context.Context) (DB, error) {
	tx, ok := db.raw.(*sql.Tx)
	if ok {
		return NewDB(tx), nil
	}
	r, ok := db.raw.(*sql.DB)
	if ok {
		tx, err := r.BeginTx(ctx, nil)
		if err != nil {
			debugf("tx", "begin tx failed: %v", tx)
			return nil, err
		}
		debugf("tx", "begin tx success")
		return NewDB(tx), nil
	}
	return nil, newError("tx", "invalid db type")
}

func (db *db) Rollback() error {
	tx, ok := db.raw.(*sql.Tx)
	if ok {
		if err := tx.Rollback(); err != nil {
			debugf("tx", "rollback tx failed: %v", err)
			return err
		}
		debugf("tx", "rollback tx success")
		return nil
	}
	return nil
}

func (db *db) Commit() error {
	tx, ok := db.raw.(*sql.Tx)
	if ok {
		if err := tx.Commit(); err != nil {
			debugf("tx", "commit tx failed: %v", err)
			return err
		}
		debugf("tx", "commit tx success")
		return nil
	}
	return nil
}

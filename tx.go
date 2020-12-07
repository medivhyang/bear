package bear

import (
	"context"
	"database/sql"
)

type (
	txContextKey    struct{}
	rawDBContextKey struct{}
)

var (
	txContextKeySingleton    = txContextKey{}
	rawDBContextKeySingleton = rawDBContextKey{}
)

func PutRawDB(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, rawDBContextKeySingleton, db)
}

func GetRawDB(ctx context.Context) *sql.DB {
	db, ok := ctx.Value(rawDBContextKeySingleton).(*sql.DB)
	if ok {
		return db
	}
	return nil
}

func MustGetRawDB(ctx context.Context) DB {
	db := GetRawDB(ctx)
	if db == nil {
		panic(ErrNoDB)
	}
	return nil
}

func PutTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKeySingleton, tx)
}

func GetTx(ctx context.Context) *sql.Tx {
	tx, ok := ctx.Value(txContextKeySingleton).(*sql.Tx)
	if ok {
		return tx
	}
	return nil
}

func GetDB(ctx context.Context) DB {
	if tx := GetTx(ctx); tx != nil {
		return tx
	}
	if db := GetRawDB(ctx); db != nil {
		return db
	}
	return nil
}

func MustGetDB(ctx context.Context) DB {
	db := GetDB(ctx)
	if db == nil {
		panic(ErrNoDB)
	}
	return db
}

func BeginTx(ctx context.Context, options ...*sql.TxOptions) (context.Context, error) {
	if tx := GetTx(ctx); tx != nil {
		return ctx, nil
	}
	db := GetRawDB(ctx)
	if db == nil {
		return ctx, ErrNoDB
	}
	var finalOptions *sql.TxOptions
	if len(options) > 0 {
		finalOptions = options[0]
	}
	tx, err := db.BeginTx(ctx, finalOptions)
	if err != nil {
		return ctx, err
	}
	return PutTx(ctx, tx), nil
}

func CommitTx(ctx context.Context) error {
	if tx := GetTx(ctx); tx != nil {
		return tx.Commit()
	}
	return nil
}

func RollbackTx(ctx context.Context) error {
	if tx := GetTx(ctx); tx != nil {
		return tx.Rollback()
	}
	return nil
}

func Tx(ctx context.Context, f func(context.Context) error, options ...*sql.TxOptions) error {
	ctx, err := BeginTx(ctx, options...)
	if err != nil {
		return err
	}
	if err := f(ctx); err != nil {
		if err := RollbackTx(ctx); err != nil {
			return err
		}
		return err
	}
	return CommitTx(ctx)
}

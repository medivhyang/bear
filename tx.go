package bear

import (
	"context"
	"database/sql"
)

func DoTx(ctx context.Context, db DB, f func(tx Tx) error, opts ...*sql.TxOptions) error {
	var finalOpts *sql.TxOptions
	if len(opts) > 0 {
		finalOpts = opts[0]
	}
	tx, err := db.BeginTx(ctx, finalOpts)
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return nil
	}
	return nil
}

type txKey struct{}

var txKeySingleton = txKey{}

func PutTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txKeySingleton, tx)
}

func GetTx(ctx context.Context) Tx {
	v, ok := ctx.Value(txKeySingleton).(Tx)
	if !ok {
		return nil
	}
	return v
}

package bear

import (
	"context"
	"database/sql"
)

func BeginTx(ctx context.Context, db Tx, f func(db DB) error, opts ...*sql.TxOptions) error {
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

package bear

import (
	"context"
	"database/sql"
)

func Tx(ctx context.Context, db DB, f func(tx *sql.Tx) error, options ...*sql.TxOptions) error {
	var finalOptions *sql.TxOptions
	if len(options) > 0 {
		finalOptions = options[0]
	}
	tx, err := db.BeginTx(ctx, finalOptions)
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

package bear

import (
	"database/sql"
)

type Result struct {
	Raw sql.Result
}

func WrapResult(result sql.Result) *Result {
	return &Result{Raw: result}
}

func (r *Result) LastInsertId() (int64, error) {
	return r.Raw.LastInsertId()
}

func (r *Result) RowsAffected() (int64, error) {
	return r.Raw.RowsAffected()
}

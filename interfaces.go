package bear

import (
	"context"
	"database/sql"
)

type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type WithContextExectutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Querier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type WithContextQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}


package bear

import (
	"context"
	"fmt"
	"strings"
)

type batchInsertBuilder struct {
	table   string
	rows    [][]Template
	include map[string]bool
	exclude map[string]bool
}

func BatchInsert(table string, rows []map[string]interface{}) *batchInsertBuilder {
	result := &batchInsertBuilder{table: table}
	for _, row := range rows {
		var columns []Template
		for k, v := range row {
			columns = append(columns, NewTemplate(k, v))
		}
		if len(columns) > 0 {
			result.rows = append(result.rows, columns)
		}
	}
	return result
}

func (b *batchInsertBuilder) Include(names ...string) *batchInsertBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *batchInsertBuilder) Exclude(names ...string) *batchInsertBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
	return b
}

func (b *batchInsertBuilder) Build() Template {
	var (
		rows    = b.finalRows()
		first   = true
		names   []string
		holders []string
		values  []interface{}
	)
	if len(rows) == 0 {
		return Template{}
	}
	for _, row := range rows {
		for _, column := range row {
			if first {
				names = append(names, column.Format)
				holders = append(holders, "?")
			}
			values = append(values, column.Values...)
		}
		first = false
	}
	var holderMatrix []string
	for i := 0; i < len(rows); i++ {
		holderMatrix = append(holderMatrix, "("+strings.Join(holders, ",")+")")
	}
	format := fmt.Sprintf("insert into %s(%s) values(%s)",
		b.table,
		strings.Join(names, ","),
		strings.Join(holderMatrix, ""),
	)
	return NewTemplate(format, values...)
}

func (b *batchInsertBuilder) Execute(exectutor Executor) (*Result, error) {
	return b.Build().Execute(exectutor)
}

func (b *batchInsertBuilder) ExecuteContext(ctx context.Context, exectutor WithContextExectutor) (*Result, error) {
	return b.Build().ExecuteContext(ctx, exectutor)
}

func (b *batchInsertBuilder) finalRows() [][]Template {
	var includedRows [][]Template
	if len(b.include) > 0 {
		for _, row := range b.rows {
			var columns []Template
			for _, column := range row {
				if b.include[column.Format] {
					columns = append(columns, column)
				}
			}
			if len(columns) > 0 {
				includedRows = append(includedRows, columns)
			}
		}
	} else {
		includedRows = b.rows
	}

	var excludedRows [][]Template
	if len(b.exclude) > 0 {
		for _, row := range b.rows {
			var columns []Template
			for _, column := range row {
				if !b.exclude[column.Format] {
					columns = append(columns, column)
				}
			}
			if len(columns) > 0 {
				excludedRows = append(excludedRows, columns)
			}
		}
	} else {
		excludedRows = includedRows
	}

	return excludedRows
}

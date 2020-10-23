package bear

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type BatchInsertBuilder struct {
	table   string
	rows    [][]Template
	include map[string]bool
	exclude map[string]bool
}

func NewBatchInsertBuilder() *BatchInsertBuilder {
	return &BatchInsertBuilder{}
}

func BatchInsert(table string, rows []map[string]interface{}) *BatchInsertBuilder {
	return NewBatchInsertBuilder().BatchInsert(table, rows)
}

func BatchInsertStruct(table string, structs interface{}) *BatchInsertBuilder {
	return NewBatchInsertBuilder().BatchInsertStruct(table, structs)
}

func (b *BatchInsertBuilder) BatchInsert(table string, rows []map[string]interface{}) *BatchInsertBuilder {
	result := &BatchInsertBuilder{table: table}
	if len(rows) == 0 {
		return result
	}
	var keys []string
	for k := range rows[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, row := range rows {
		var columns []Template
		for _, k := range keys {
			columns = append(columns, NewTemplate(k, row[k]))
		}
		if len(columns) > 0 {
			result.rows = append(result.rows, columns)
		}
	}
	return result
}

func (b *BatchInsertBuilder) BatchInsertStruct(table string, structs interface{}) *BatchInsertBuilder {
	v := reflect.ValueOf(structs)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		panic("batch insert struct: require struct slice")
	}
	if v.Len() == 0 {
		return &BatchInsertBuilder{}
	}
	var rows []map[string]interface{}
	for i := 0; i < v.Len(); i++ {
		row := toColumnValueMapFromStruct(v.Index(i), true)
		rows = append(rows, row)
	}
	return BatchInsert(table, rows)
}

func (b *BatchInsertBuilder) Table(name string) *BatchInsertBuilder {
	b.table = name
	return b
}

func (b *BatchInsertBuilder) Include(names ...string) *BatchInsertBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *BatchInsertBuilder) Exclude(names ...string) *BatchInsertBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
	return b
}

func (b *BatchInsertBuilder) Build() Template {
	rows := b.finalRows()
	if len(rows) == 0 {
		return Template{}
	}
	var (
		names   []string
		holders []string
	)
	for _, column := range rows[0] {
		names = append(names, column.Format)
		holders = append(holders, "?")
	}
	var (
		holderMatrix []string
		values       []interface{}
	)
	for _, row := range rows {
		holderMatrix = append(holderMatrix, "("+strings.Join(holders, ",")+")")
		for _, column := range row {
			values = append(values, column.Values...)
		}
	}
	format := fmt.Sprintf("insert into %s(%s) values%s",
		b.table,
		strings.Join(names, ","),
		strings.Join(holderMatrix, ","),
	)
	return NewTemplate(format, values...)
}

func (b *BatchInsertBuilder) Execute(db DB) (*Result, error) {
	return b.Build().Execute(db)
}

func (b *BatchInsertBuilder) ExecuteContext(ctx context.Context, db DB) (*Result, error) {
	return b.Build().ExecuteContext(ctx, db)
}

func (b *BatchInsertBuilder) finalRows() [][]Template {
	if len(b.rows) == 0 {
		return nil
	}

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
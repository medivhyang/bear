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

func BatchInsert(table string, args ...interface{}) *BatchInsertBuilder {
	return NewBatchInsertBuilder().BatchInsert(table, args...)
}

func (b *BatchInsertBuilder) BatchInsert(table string, args ...interface{}) *BatchInsertBuilder {
	if len(args) == 0 {
		return b
	}
	firstArg := args[0]
	switch v := firstArg.(type) {
	case map[string]interface{}:
		var items []map[string]interface{}
		for _, arg := range args {
			items = append(items, arg.(map[string]interface{}))
		}
		b.batchInsertMaps(table, items...)
	case []map[string]interface{}:
		b.batchInsertMaps(table, v...)
	default:
		firstArgValue := reflect.ValueOf(firstArg)
		for firstArgValue.Kind() == reflect.Ptr {
			firstArgValue = firstArgValue.Elem()
		}
		switch firstArgValue.Kind() {
		case reflect.Struct:
			b.batchInsertStructs(table, args)
		case reflect.Slice:
			if firstArgValue.Type().Elem().Kind() != reflect.Struct {
				panic("bear: batch insert builder batch insert: unsupported args type")
			}
			b.batchInsertStructs(table, toInterfaceSlice(firstArg))
		default:
			panic("bear: batch insert builder batch insert: unsupported args type")
		}
	}
	return b
}

func (b *BatchInsertBuilder) batchInsertMaps(table string, rows ...map[string]interface{}) *BatchInsertBuilder {
	if len(rows) == 0 {
		return b
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
			b.rows = append(b.rows, columns)
		}
	}
	return b
}

func (b *BatchInsertBuilder) batchInsertStructs(table string, structs []interface{}) *BatchInsertBuilder {
	if len(structs) == 0 {
		return b
	}
	var rows []map[string]interface{}
	for _, aStruct := range structs {
		row := mapStructToColumns(reflect.ValueOf(aStruct), true)
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

func (b *BatchInsertBuilder) Exec(ctx context.Context, db DB) (*Result, error) {
	return b.Build().Exec(ctx, db)
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

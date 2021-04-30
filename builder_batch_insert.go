package bear

import (
	"fmt"
	"github.com/medivhyang/ice"
	"sort"
	"strings"
)

type BatchInsertBuilder struct {
	Dialect string
	Table   string
	Rows    []Templates
}

func NewBatchInsertBuilder(dialect ...string) *BatchInsertBuilder {
	if len(dialect) == 0 {
		return &BatchInsertBuilder{}
	}
	return &BatchInsertBuilder{Dialect: dialect[0]}
}

func (b *BatchInsertBuilder) Apply(options ...BatchInsertBuilderOptionFunc) *BatchInsertBuilder {
	for _, option := range options {
		if option == nil {
			continue
		}
		option(b)
	}
	return b
}

func (b *BatchInsertBuilder) Build() Template {
	if len(b.Rows) == 0 {
		return Template{}
	}
	firstRow := b.Rows[0]
	columns := make([]string, 0, len(firstRow))
	rowHolders := repeatString("?", len(firstRow))
	holders := make([]string, 0, len(b.Rows))
	values := make([]interface{}, 0, len(b.Rows)*len(firstRow))
	for _, c := range firstRow {
		columns = append(columns, c.Format)
		holders = append(holders, "("+strings.Join(rowHolders, ", ")+")")
		values = append(values, c.Values...)
	}
	return NewTemplate(fmt.Sprintf("insert into %s(%s) values%s",
		b.Table,
		strings.Join(columns, ", "),
		strings.Join(holders, ", "),
	), values...)
}

func (b *BatchInsertBuilder) Append(rows ...Templates) *BatchInsertBuilder {
	b.Rows = append(b.Rows, rows...)
	return b
}

func (b *BatchInsertBuilder) AppendMap(rows ...map[string]interface{}) *BatchInsertBuilder {
	if len(rows) == 0 {
		return b
	}
	firstRow := rows[0]
	kk := make([]string, 0, len(firstRow))
	for k := range firstRow {
		kk = append(kk, k)
	}
	sort.Strings(kk)
	for _, r := range rows {
		tt := NewTemplates()
		for _, k := range kk {
			tt.Appendf(k, r[k])
		}
		if len(tt) == 0 {
			continue
		}
		b.Rows = append(b.Rows, tt)
	}
	return b
}

func (b *BatchInsertBuilder) AppendStruct(rows ...interface{}) *BatchInsertBuilder {
	mm := make([]map[string]interface{}, len(rows))
	for _, r := range rows {
		mm = append(mm, ice.ParseStructToMap(r))
	}
	return b.AppendMap(mm...)
}

type BatchInsertBuilderOptionFunc func(b *BatchInsertBuilder)

func BatchInsert(table string, rows ...Templates) BatchInsertBuilderOptionFunc {
	return func(b *BatchInsertBuilder) {
		b.Table = table
		b.Append(rows...)
	}
}

func BatchInsertMap(table string, rows ...map[string]interface{}) BatchInsertBuilderOptionFunc {
	return func(b *BatchInsertBuilder) {
		b.Table = table
		b.AppendMap(rows...)
	}
}

func BatchInsertStruct(table string, rows ...interface{}) BatchInsertBuilderOptionFunc {
	return func(b *BatchInsertBuilder) {
		b.Table = table
		b.AppendStruct(rows...)
	}
}

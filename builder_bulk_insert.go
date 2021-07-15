package bear

import (
	"fmt"
	"strings"

	"github.com/medivhyang/duck/reflectutil"
)

type BulkInsertBuilder struct {
	dialect string
	table   string
	columns []string
	values  [][]interface{}
	err     error
}

func NewBulkInsertBuilder() *BulkInsertBuilder {
	return &BulkInsertBuilder{}
}

func (b *BulkInsertBuilder) Build() (Template, error) {
	if b.err != nil {
		return Template{}, b.err
	}

	holders := make([]string, 0, len(b.values))
	for i := 0; i < len(b.values); i++ {
		holders = append(holders, "("+strings.Join(repeatString("?", len(b.columns)), ", ")+")")
	}

	flattenValues := make([]interface{}, 0, len(b.columns)*len(b.values))
	for _, rowValues := range b.values {
		flattenValues = append(flattenValues, rowValues...)
	}

	t := NewTemplate(fmt.Sprintf("insert into %s(%s) values%s",
		b.table,
		strings.Join(b.columns, ", "),
		strings.Join(holders, ", "),
	), flattenValues...)

	return t, nil
}

func (b *BulkInsertBuilder) Dialect(name string) *BulkInsertBuilder {
	b.dialect = name
	return b
}

func (b *BulkInsertBuilder) Table(name string) *BulkInsertBuilder {
	b.table = name
	return b
}

func (b *BulkInsertBuilder) Columns(cc ...string) *BulkInsertBuilder {
	b.columns = cc
	return b
}

func (b *BulkInsertBuilder) Append(values ...[]interface{}) *BulkInsertBuilder {
	b.values = append(b.values, values...)
	return b
}

func (b *BulkInsertBuilder) AppendMap(mm ...map[string]interface{}) *BulkInsertBuilder {
	for _, m := range mm {
		rowValues := make([]interface{}, 0, len(b.columns))
		for _, c := range b.columns {
			v, ok := m[c]
			if ok {
				rowValues = append(rowValues, v)
			} else {
				b.err = newError("bulk builder append maps", "missing value for %s", c)
				return b
			}
		}
		b.values = append(b.values, rowValues)
	}
	return b
}

func (b *BulkInsertBuilder) AppendStruct(ii ...interface{}) *BulkInsertBuilder {
	var mm []map[string]interface{}
	for _, i := range ii {
		m := reflectutil.ParseStructToMap(i)
		mm = append(mm, m)
	}
	b.AppendMap(mm...)
	return b
}

package bear

import (
	"fmt"
	"github.com/medivhyang/duck/ice"
	"github.com/medivhyang/duck/slices"
	"sort"
	"strings"
)

type BulkBuilder struct {
	dialect string
	table   string
	rows    []Templates
}

func NewBulkBuilder() *BulkBuilder {
	return &BulkBuilder{}
}

func (b *BulkBuilder) Build() Template {
	if len(b.rows) == 0 {
		return Template{}
	}
	firstRow := b.rows[0]
	columns := make([]string, 0, len(firstRow))
	rowHolders := repeatString("?", len(firstRow))
	holders := make([]string, 0, len(b.rows))
	values := make([]interface{}, 0, len(b.rows)*len(firstRow))
	for _, c := range firstRow {
		columns = append(columns, c.Format)
		holders = append(holders, "("+strings.Join(rowHolders, ", ")+")")
		values = append(values, c.Values...)
	}
	return NewTemplate(fmt.Sprintf("insert into %s(%s) values%s",
		b.table,
		strings.Join(columns, ", "),
		strings.Join(holders, ", "),
	), values...)
}

func (b *BulkBuilder) Dialect(name string) *BulkBuilder {
	b.dialect = name
	return b
}

func (b *BulkBuilder) Table(name string) *BulkBuilder {
	b.table = name
	return b
}

func (b *BulkBuilder) Append(row Templates) *BulkBuilder {
	b.rows = append(b.rows, row)
	return b
}

func (b *BulkBuilder) AppendMap(row map[string]interface{}) *BulkBuilder {
	keys := make([]string, 0, len(row))
	for key := range row {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	tt := NewTemplates()
	for _, k := range keys {
		tt.Appendf(k, row[k])
	}
	b.rows = append(b.rows, tt)

	return b
}

func (b *BulkBuilder) AppendStruct(i interface{}, ignoreFields ...string) *BulkBuilder {
	m := ice.ParseStructToMap(i)
	m2 := make(map[string]interface{}, len(m))
	for name, value := range m {
		if slices.ContainStrings(ignoreFields, name) {
			continue
		}
		m2[name] = value
	}
	b.AppendMap(m2)
	return b
}

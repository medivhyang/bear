package bear

import (
	"fmt"
	"github.com/medivhyang/ice"
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

func (b *BulkBuilder) Append(rows ...Templates) *BulkBuilder {
	b.rows = append(b.rows, rows...)
	return b
}

func (b *BulkBuilder) AppendMaps(rows ...map[string]interface{}) *BulkBuilder {
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
		b.rows = append(b.rows, tt)
	}
	return b
}

func (b *BulkBuilder) AppendStructs(rows ...interface{}) *BulkBuilder {
	mm := make([]map[string]interface{}, len(rows))
	for _, r := range rows {
		mm = append(mm, ice.ParseStructToMap(r))
	}
	return b.AppendMaps(mm...)
}

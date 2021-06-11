package bear

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/medivhyang/duck/ice"
	"github.com/medivhyang/duck/slices"
)

type Action string

const (
	actionSelect Action = "select"
	actionInsert Action = "insert"
	actionUpdate Action = "update"
	actionDelete Action = "delete"
)

type Builder struct {
	action   Action
	dialect  string
	table    Template
	columns  Templates
	joins    Templates
	where    Conditions
	orderBy  []string
	paging   Template
	groupBy  []string
	having   Conditions
	distinct bool
}

func NewBuilder(options ...BuilderOptionFunc) *Builder {
	return new(Builder).Apply(options...)
}

func (b *Builder) Apply(options ...BuilderOptionFunc) *Builder {
	for _, option := range options {
		if option == nil {
			continue
		}
		option(b)
	}
	return b
}

func (b *Builder) Build() Template {
	t := Template{}
	switch b.action {
	case actionSelect:
		if b.distinct {
			t = t.Appendf(fmt.Sprintf("select distinct %s from %s",
				strings.Join(b.columns.Formats(), ","),
				GetDialect(b.dialect).Quote(b.table.Format),
			), append(append([]interface{}{}, b.columns.Values()...), b.table.Values...))
		} else {
			t = t.Appendf(fmt.Sprintf("select %s from %s",
				strings.Join(b.columns.Formats(), ","),
				b.table,
			), b.columns.Values()...)
		}
		t = t.Append(b.joins.Join(" ", "", ""))
		if len(b.where) > 0 {
			t = t.Appendf(" where ").Append(b.where.And())
		}
		if len(b.groupBy) > 0 {
			t = t.Appendf(fmt.Sprintf(" group by %s", strings.Join(b.groupBy, ",")))
		}
		if len(b.having) > 0 {
			t = t.Appendf(" having ").Append(b.having.And())
		}
		if len(b.orderBy) > 0 {
			t = t.Appendf(fmt.Sprintf(" order by %s", strings.Join(b.orderBy, ", ")))
		}
		if len(b.paging.Format) > 0 {
			t = t.Appendf(" ").Append(b.paging)
		}
	case actionInsert:
		t = t.Appendf(fmt.Sprintf("insert into %s(%s) values(%s)",
			b.table,
			b.columns.Formats(),
			strings.Join(repeatString("?", len(b.columns)), ","),
		), b.columns.Values()...)
		if len(b.where) > 0 {
			t = t.Appendf(" where ").Append(b.where.And())
		}
	case actionUpdate:
		pairs := make([]string, 0, len(b.columns))
		for _, c := range b.columns {
			pairs = append(pairs, fmt.Sprintf("%s = ?", GetDialect(b.dialect).Quote(c.Format)))
		}
		t.Appendf(fmt.Sprintf("update %s set %s",
			b.table,
			strings.Join(pairs, ","),
		), b.columns.Values()...)
		if len(b.where) > 0 {
			t = t.Appendf(" where ").Append(b.where.And())
		}
	case actionDelete:
		t = t.Appendf(fmt.Sprintf("delete from %s", b.table))
		if len(b.where) > 0 {
			t = t.Appendf(" where ").Append(b.where.And())
		}
	}
	return t
}

func (b *Builder) Dialect(d string) *Builder {
	b.dialect = d
	return b
}

func (b *Builder) Select(table string, columns ...string) *Builder {
	b.action = actionSelect
	b.table = NewTemplate(table)
	for _, c := range columns {
		b.columns = append(b.columns, NewTemplate(c))
	}
	return b
}

func (b *Builder) SelectStruct(table string, i interface{}, ignoreFields ...string) *Builder {
	cc := parseStructColumns(i, ignoreFields...)
	b.action = actionSelect
	b.table = NewTemplate(table)
	for _, c := range cc {
		b.columns = append(b.columns, NewTemplate(c))
	}
	return b
}

func (b *Builder) Insert(table string, columns map[string]interface{}) *Builder {
	b.action = actionInsert
	b.table = NewTemplate(table)
	for name, value := range columns {
		b.columns = append(b.columns, NewTemplate(name, value))
	}
	return b
}

func (b *Builder) InsertStruct(table string, i interface{}, ignoreZeroValue bool, ignoreFields ...string) *Builder {
	b.action = actionInsert
	b.table = NewTemplate(table)
	m := ice.ParseStructToMap(i)
	for name, value := range m {
		if ignoreZeroValue && reflect.ValueOf(value).IsZero() {
			continue
		}
		if slices.ContainStrings(ignoreFields, name) {
			continue
		}
		b.columns = append(b.columns, NewTemplate(name, value))
	}
	return b
}

func (b *Builder) Update(table string, columns map[string]interface{}) *Builder {
	b.action = actionUpdate
	b.table = NewTemplate(table)
	for name, value := range columns {
		b.columns = append(b.columns, NewTemplate(name, value))
	}
	return b
}

func (b *Builder) UpdateStruct(table string, i interface{}, ignoreZeroValue bool, ignoreFields ...string) *Builder {
	b.action = actionUpdate
	b.table = NewTemplate(table)
	m := ice.ParseStructToMap(i)
	for name, value := range m {
		if ignoreZeroValue && reflect.ValueOf(value).IsZero() {
			continue
		}
		if slices.ContainStrings(ignoreFields, name) {
			continue
		}
		b.columns = append(b.columns, NewTemplate(name, value))
	}
	return b
}

func (b *Builder) Delete(table string) *Builder {
	b.action = actionDelete
	b.table = NewTemplate(table)
	return b
}

func (b *Builder) Where(format string, values ...interface{}) *Builder {
	b.where.Appendf(format, values...)
	return b
}

func (b *Builder) OrderBy(fields ...string) *Builder {
	b.orderBy = append(b.orderBy, fields...)
	return b
}

func (b *Builder) Paging(page, size int) *Builder {
	b.paging = NewTemplate("limit ?,?", (page-1)*size, size)
	return b
}

func (b *Builder) GroupBy(fields ...string) *Builder {
	b.groupBy = append(b.groupBy, fields...)
	return b
}

func (b *Builder) Having(format string, values ...interface{}) *Builder {
	b.having.Appendf(format, values...)
	return b
}

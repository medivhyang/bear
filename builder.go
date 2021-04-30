package bear

import (
	"fmt"
	"strings"
)

type Action string

const (
	ActionSelect Action = "select"
	ActionInsert Action = "insert"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Builder struct {
	Action   Action
	Dialect  string
	Table    Template
	Columns  Templates
	Joins    Templates
	Where    Conditions
	OrderBy  []string
	Paging   Template
	GroupBy  []string
	Having   Conditions
	Distinct bool
}

func NewBuilder(dialect ...string) *Builder {
	if len(dialect) == 0 {
		return &Builder{}
	}
	return &Builder{Dialect: dialect[0]}
}

func B(options ...BuilderOptionFunc) Template {
	return NewBuilder().Apply(options...).Build()
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
	switch b.Action {
	case ActionSelect:
		if b.Distinct {
			t = t.Appendf(fmt.Sprintf("select distinct %s from %s",
				strings.Join(b.Columns.Formats(), ","),
				GetDialect(b.Dialect).Quote(b.Table.Format),
			), append(append([]interface{}{}, b.Columns.Values()...), b.Table.Values...))
		} else {
			t = t.Appendf(fmt.Sprintf("select %s from %s",
				strings.Join(b.Columns.Formats(), ","),
				b.Table,
			), b.Columns.Values()...)
		}
		t = t.Append(b.Joins.Join(" ", "", ""))
		if len(b.Where) > 0 {
			t = t.Appendf(" where ").Append(b.Where.And())
		}
		if len(b.GroupBy) > 0 {
			t = t.Appendf(fmt.Sprintf(" group by %s", strings.Join(b.GroupBy, ",")))
		}
		if len(b.Having) > 0 {
			t = t.Appendf(" having ").Append(b.Having.And())
		}
		if len(b.OrderBy) > 0 {
			t = t.Appendf(fmt.Sprintf(" order by %s", strings.Join(b.OrderBy, ", ")))
		}
		if len(b.Paging.Format) > 0 {
			t = t.Appendf(" ").Append(b.Paging)
		}
	case ActionInsert:
		t = t.Appendf(fmt.Sprintf("insert into %s(%s) values(%s)",
			b.Table,
			b.Columns.Formats(),
			strings.Join(repeatString("?", len(b.Columns)), ","),
		), b.Columns.Values()...)
		if len(b.Where) > 0 {
			t = t.Appendf(" where ").Append(b.Where.And())
		}
	case ActionUpdate:
		pairs := make([]string, 0, len(b.Columns))
		for _, c := range b.Columns {
			pairs = append(pairs, fmt.Sprintf("%s = ?", GetDialect(b.Dialect).Quote(c.Format)))
		}
		t.Appendf(fmt.Sprintf("update %s set %s",
			b.Table,
			strings.Join(pairs, ","),
		), b.Columns.Values()...)
		if len(b.Where) > 0 {
			t = t.Appendf(" where ").Append(b.Where.And())
		}
	case ActionDelete:
		t = t.Appendf(fmt.Sprintf("delete from %s", b.Table))
		if len(b.Where) > 0 {
			t = t.Appendf(" where ").Append(b.Where.And())
		}
	}
	return t
}

type BuilderOptionFunc func(b *Builder)

func Select(table string, columns ...string) BuilderOptionFunc {
	return func(b *Builder) {
		b.Action = ActionSelect
		b.Table = NewTemplate(table)
		for _, c := range columns {
			b.Columns = append(b.Columns, NewTemplate(c))
		}
	}
}

func SelectStruct(table string, i interface{}, ignoreColumns ...string) BuilderOptionFunc {
	cc := parseStructColumns(i, ignoreColumns...)
	return func(b *Builder) {
		b.Action = ActionSelect
		b.Table = NewTemplate(table)
		for _, c := range cc {
			b.Columns = append(b.Columns, NewTemplate(c))
		}
	}
}

func Insert(table string, columns map[string]interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.Action = ActionInsert
		b.Table = NewTemplate(table)
		for name, value := range columns {
			b.Columns = append(b.Columns, NewTemplate(name, value))
		}
	}
}

func Update(table string, columns map[string]interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.Action = ActionUpdate
		b.Table = NewTemplate(table)
		for name, value := range columns {
			b.Columns = append(b.Columns, NewTemplate(name, value))
		}
	}
}

func Delete(table string) BuilderOptionFunc {
	return func(b *Builder) {
		b.Action = ActionDelete
		b.Table = NewTemplate(table)
	}
}

func Where(format string, values ...interface{}) BuilderOptionFunc {
	t := NewTemplate(format, values)
	return func(b *Builder) {
		b.Where = append(b.Where, t)
	}
}

func OrderBy(fields ...string) BuilderOptionFunc {
	return func(b *Builder) {
		b.OrderBy = append(b.OrderBy, fields...)
	}
}

func Paging(page, size int) BuilderOptionFunc {
	return func(b *Builder) {
		b.Paging = NewTemplate("limit ?,?", (page-1)*size, size)
	}
}

func GroupBy(fields ...string) BuilderOptionFunc {
	return func(b *Builder) {
		b.GroupBy = append(b.GroupBy, fields...)
	}
}

func Having(format string, values ...interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.Having.Appendf(format, values...)
	}
}

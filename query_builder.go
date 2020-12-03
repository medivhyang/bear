package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type QueryBuilder struct {
	dialect  string
	table    string
	distinct bool
	columns  []Template
	include  map[string]bool
	exclude  map[string]bool
	joins    []Template
	where    Condition
	groupBy  []string
	having   Condition
	orderBy  []string
	paging   Template
}

func NewQueryBuilder(dialect ...string) *QueryBuilder {
	b := &QueryBuilder{}
	if len(dialect) > 0 {
		b.Dialect(dialect[0])
	}
	return b
}

func Select(table string, columns ...interface{}) *QueryBuilder {
	return NewQueryBuilder().Select(table, columns...)
}

func (b *QueryBuilder) Select(table string, columns ...interface{}) *QueryBuilder {
	return b.Table(table).Columns(columns...)
}

func (b *QueryBuilder) Dialect(name string) *QueryBuilder {
	b.dialect = name
	return b
}

func (b *QueryBuilder) Table(name string) *QueryBuilder {
	b.table = name
	return b
}

func (b *QueryBuilder) Distinct(value bool) *QueryBuilder {
	b.distinct = value
	return b
}

func (b *QueryBuilder) Columns(columns ...interface{}) *QueryBuilder {
	if len(columns) == 0 {
		return b
	}
	firstColumn := columns[0]
	switch firstColumn.(type) {
	case string:
		var items []string
		for _, c := range columns {
			items = append(items, c.(string))
		}
		b.stringColumns(items...)
	case Template:
		var items []Template
		for _, c := range columns {
			items = append(items, c.(Template))
		}
		b.templateColumns(items...)
	default:
		firstColumnValue := reflect.ValueOf(firstColumn)
		for firstColumnValue.Kind() == reflect.Ptr {
			firstColumnValue = firstColumnValue.Elem()
		}
		switch firstColumnValue.Kind() {
		case reflect.Struct:
			b.structColumns(firstColumn)
		default:
			panic("bear: columns: unsupported args type")
		}
	}
	return b
}

func (b *QueryBuilder) stringColumns(columns ...string) *QueryBuilder {
	for _, column := range columns {
		b.columns = append(b.columns, NewTemplate(column))
	}
	return b
}

func (b *QueryBuilder) templateColumns(columns ...Template) *QueryBuilder {
	b.columns = append(b.columns, columns...)
	return b
}

func (b *QueryBuilder) structColumns(aStruct interface{}) *QueryBuilder {
	names := parseFields(reflect.TypeOf(aStruct)).columnNames()
	for _, name := range names {
		b.columns = append(b.columns, NewTemplate(name))
	}
	return b
}

func (b *QueryBuilder) Include(names ...string) *QueryBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *QueryBuilder) Exclude(names ...string) *QueryBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
	return b
}

func (b *QueryBuilder) Join(args ...interface{}) *QueryBuilder {
	if len(args) == 0 {
		return b
	}
	firstArg := args[0]
	switch firstArg.(type) {
	case string:
		b.joins = append(b.joins, Template{Format: args[0].(string), Values: args[1:]})
	case Template:
		var items []Template
		for _, arg := range args {
			items = append(items, arg.(Template))
		}
		b.joins = append(b.joins, items...)
	default:
		panic("bear: query builder join: unsupported args type")
	}
	return b
}

func (b *QueryBuilder) Where(args ...interface{}) *QueryBuilder {
	b.where = b.where.Append(args...)
	return b
}

func (b *QueryBuilder) Having(args ...interface{}) *QueryBuilder {
	b.having.Append(args...)
	return b
}

func (b *QueryBuilder) GroupBy(names ...string) *QueryBuilder {
	b.groupBy = append(b.groupBy, names...)
	return b
}

func (b *QueryBuilder) OrderBy(names ...string) *QueryBuilder {
	b.orderBy = append(b.orderBy, names...)
	return b
}

func (b *QueryBuilder) Limit(offset int, limit int) *QueryBuilder {
	b.paging = NewTemplate("limit ?,?", offset, limit)
	return b
}

func (b *QueryBuilder) Paging(args ...interface{}) *QueryBuilder {
	if len(args) == 0 {
		return b
	}
	firstArg := args[0]
	switch firstArg.(type) {
	case string:
		b.paging = NewTemplate(firstArg.(string), args[1:]...)
	case Template:
		b.paging = firstArg.(Template)
	default:
		panic("bear: query builder: unsupported args type")
	}
	return b
}

func (b *QueryBuilder) As(alias string) Template {
	t := b.Build()
	if len(alias) == 0 {
		return t
	}
	return t.WrapBracket().Append(fmt.Sprintf(" as %s", alias))
}

func (b *QueryBuilder) Build() Template {
	columns := b.finalColumns()
	if len(columns) == 0 {
		columns = []Template{NewTemplate("*")}
	}

	result := Template{}

	var formats []string
	{
		for _, c := range columns {
			formats = append(formats, c.Format)
			result.Values = append(result.Values, c.Values...)
		}
		if b.distinct {
			result.Format = fmt.Sprintf("select distinct %s from %s",
				strings.Join(formats, ","),
				b.table,
			)
		} else {
			result.Format = fmt.Sprintf("select %s from %s",
				strings.Join(formats, ","),
				b.table,
			)
		}
	}

	for _, item := range b.joins {
		result = result.Join(item, " ")
	}

	where := b.where.Build()
	if !where.IsEmptyOrWhitespace() {
		result = result.Join(where, " where ")
	}

	if len(b.groupBy) > 0 {
		result = result.Append(fmt.Sprintf(" group by %s", strings.Join(b.groupBy, ",")))
	}

	having := b.having.Build()
	if !having.IsEmptyOrWhitespace() {
		result = result.Join(having, " having ")
	}

	if len(b.orderBy) > 0 {
		result = result.Append(fmt.Sprintf(" order by %s", strings.Join(b.orderBy, ",")))
	}

	if !b.paging.IsEmptyOrWhitespace() {
		result = result.Join(b.paging, " ")
	}

	return result
}

func (b *QueryBuilder) Query(ctx context.Context, db DB, value interface{}) error {
	return b.Build().Query(ctx, db, value)
}

func (b *QueryBuilder) QueryRows(ctx context.Context, db DB) (*Rows, error) {
	return b.Build().QueryRows(ctx, db)
}

func (b *QueryBuilder) finalColumns() []Template {
	var includedColumns []Template
	if len(b.include) > 0 {
		for _, column := range b.columns {
			if b.include[column.Format] {
				includedColumns = append(includedColumns, column)
			}
		}
	} else {
		includedColumns = b.columns
	}
	var excludedColumns []Template
	if len(b.exclude) > 0 {
		for _, column := range includedColumns {
			if !b.exclude[column.Format] {
				excludedColumns = append(excludedColumns, column)
			}
		}
	} else {
		excludedColumns = includedColumns
	}
	return excludedColumns
}

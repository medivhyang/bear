package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type queryBuilder struct {
	dialect string
	table   string
	columns []Template
	joins   []Template
	where   Condition
	groupBy []string
	having  Condition
	orderBy []string
	paging  Template
}

func Select(table string, columns ...string) *queryBuilder {
	b := queryBuilder{table: table}
	for _, column := range columns {
		b.columns = append(b.columns, NewTemplate(column))
	}
	return &b
}

func SelectTemplate(table string, columns ...Template) *queryBuilder {
	b := queryBuilder{table: table}
	b.columns = append(b.columns, columns...)
	return &b
}

func SelectStruct(i interface{}) *queryBuilder {
	return Select(TableName(i), structFields(reflect.TypeOf(i)).columnNames()...)
}

func SelectWhere(i interface{}) *queryBuilder {
	return Select(TableName(i), structFields(reflect.TypeOf(i)).columnNames()...).WhereStruct(i)
}

func (b *queryBuilder) Dialect(name string) *queryBuilder {
	b.dialect = name
	return b
}

func (b *queryBuilder) Table(name string) *queryBuilder {
	b.table = name
	return b
}

func (b *queryBuilder) Join(format string, values ...interface{}) *queryBuilder {
	b.joins = append(b.joins, Template{Format: format, Values: values})
	return b
}

func (b *queryBuilder) JoinTemplate(templates ...Template) *queryBuilder {
	b.joins = append(b.joins, templates...)
	return b
}

func (b *queryBuilder) Where(format string, values ...interface{}) *queryBuilder {
	b.where = b.where.AppendTemplate(NewTemplate(format, values...))
	return b
}

func (b *queryBuilder) WhereTemplate(templates ...Template) *queryBuilder {
	b.where = b.where.AppendTemplate(templates...)
	return b
}

func (b *queryBuilder) WhereMap(m map[string]interface{}) *queryBuilder {
	b.where = b.where.AppendMap(m)
	return b
}

func (b *queryBuilder) WhereStruct(i interface{}) *queryBuilder {
	b.where = b.where.AppendStruct(i)
	return b
}

func (b *queryBuilder) Having(format string, values ...interface{}) *queryBuilder {
	b.having.AppendTemplate(Template{Format: format, Values: values})
	return b
}

func (b *queryBuilder) HavingTemplate(templates ...Template) *queryBuilder {
	b.having.AppendTemplate(templates...)
	return b
}

func (b *queryBuilder) HavingMap(m map[string]interface{}) *queryBuilder {
	b.where.AppendMap(m)
	return b
}

func (b *queryBuilder) HavingStruct(i interface{}) *queryBuilder {
	b.where.AppendStruct(i)
	return b
}

func (b *queryBuilder) GroupBy(names ...string) *queryBuilder {
	b.groupBy = append(b.groupBy, names...)
	return b
}

func (b *queryBuilder) OrderBy(names ...string) *queryBuilder {
	b.orderBy = append(b.orderBy, names...)
	return b
}

func (b *queryBuilder) Limit(offset int, limit int) *queryBuilder {
	b.paging = NewTemplate("limit ?,?", offset, limit)
	return b
}

func (b *queryBuilder) Paging(format string, values ...interface{}) *queryBuilder {
	b.paging = NewTemplate(format, values...)
	return b
}

func (b *queryBuilder) PagingTemplate(template Template) *queryBuilder {
	b.paging = template
	return b
}

func (b *queryBuilder) Build() Template {
	result := Template{}

	var formats []string
	for _, c := range b.columns {
		formats = append(formats, c.Format)
		result.Values = append(result.Values, c.Values...)
	}
	result.Format = fmt.Sprintf("select %s from %s",
		strings.Join(formats, ","),
		b.table,
	)

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

func (b *queryBuilder) As(alias string) Template {
	t := b.Build()
	if len(alias) == 0 {
		return t
	}
	return t.WrapBracket().Append(fmt.Sprintf(" as %s", alias))
}

func (b *queryBuilder) Query(querier Querier) (*Rows, error) {
	return b.Build().Query(querier)
}

func (b *queryBuilder) QueryContext(ctx context.Context, querier WithContextQuerier) (*Rows, error) {
	return b.Build().QueryContext(ctx, querier)
}

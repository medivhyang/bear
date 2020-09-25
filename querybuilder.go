package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type queryBuilder struct {
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

func SelectStruct(aStruct interface{}) *queryBuilder {
	return Select(TableName(aStruct)).IncludeStruct(aStruct)
}

func SelectWhere(aStruct interface{}) *queryBuilder {
	return Select(TableName(aStruct)).IncludeStruct(aStruct).WhereStruct(aStruct)
}

func (b *queryBuilder) Dialect(name string) *queryBuilder {
	b.dialect = name
	return b
}

func (b *queryBuilder) Table(name string) *queryBuilder {
	b.table = name
	return b
}

func (b *queryBuilder) Distinct(value bool) *queryBuilder {
	b.distinct = value
	return b
}

func (b *queryBuilder) Include(names ...string) *queryBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *queryBuilder) IncludeStruct(aStruct interface{}) *queryBuilder {
	names := structFields(reflect.TypeOf(aStruct)).columnNames()
	for _, name := range names {
		b.columns = append(b.columns, NewTemplate(name))
	}
	return b
}

func (b *queryBuilder) Exclude(names ...string) *queryBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
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

func (b *queryBuilder) As(alias string) Template {
	t := b.Build()
	if len(alias) == 0 {
		return t
	}
	return t.WrapBracket().Append(fmt.Sprintf(" as %s", alias))
}

func (b *queryBuilder) Build() Template {
	columns := b.finalColumns()
	if len(columns) == 0 {
		return Template{}
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

func (b *queryBuilder) Query(querier Querier) (*Rows, error) {
	return b.Build().Query(querier)
}

func (b *queryBuilder) QueryContext(ctx context.Context, querier WithContextQuerier) (*Rows, error) {
	return b.Build().QueryContext(ctx, querier)
}

func (b *queryBuilder) QueryScalar(ctx context.Context, querier WithContextQuerier, value interface{}) error {
	return b.Build().QueryScalar(ctx, querier, value)
}

func (b *queryBuilder) QueryScalarSlice(ctx context.Context, querier WithContextQuerier, values interface{}) error {
	return b.Build().QueryScalarSlice(ctx, querier, values)
}

func (b *queryBuilder) QueryMap(ctx context.Context, querier WithContextQuerier) (map[string]interface{}, error) {
	return b.Build().QueryMap(ctx, querier)
}

func (b *queryBuilder) QueryMapSlice(ctx context.Context, querier WithContextQuerier) ([]map[string]interface{}, error) {
	return b.Build().QueryMapSlice(ctx, querier)
}

func (b *queryBuilder) QueryStruct(ctx context.Context, querier WithContextQuerier, structPtr interface{}) error {
	return b.Build().QueryStruct(ctx, querier, structPtr)
}

func (b *queryBuilder) QueryStructSlice(ctx context.Context, querier WithContextQuerier, structPtr interface{}) error {
	return b.Build().QueryStructSlice(ctx, querier, structPtr)
}

func (b *queryBuilder) finalColumns() []Template {
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

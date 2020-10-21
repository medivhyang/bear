package bear

import (
	"context"
	"errors"
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

func Select(table string, columns ...string) *QueryBuilder {
	return NewQueryBuilder().Select(table, columns...)
}

func SelectTemplate(table string, columns ...Template) *QueryBuilder {
	return NewQueryBuilder().SelectTemplate(table, columns...)
}

func SelectStruct(table string, aStruct interface{}) *QueryBuilder {
	return NewQueryBuilder().SelectStruct(table, aStruct)
}

func SelectWhereStruct(table string, aStruct interface{}) *QueryBuilder {
	return NewQueryBuilder().SelectWhereStruct(table, aStruct)
}

func (b *QueryBuilder) Select(table string, columns ...string) *QueryBuilder {
	return b.Table(table).Columns(columns...)
}

func (b *QueryBuilder) SelectTemplate(table string, columns ...Template) *QueryBuilder {
	return b.Table(table).TemplateColumns(columns...)
}

func (b *QueryBuilder) SelectStruct(table string, aStruct interface{}) *QueryBuilder {
	return b.Table(table).StructColumns(aStruct)
}

func (b *QueryBuilder) SelectWhereStruct(table string, aStruct interface{}, includeZeroValue ...bool) *QueryBuilder {
	return b.Table(table).StructColumns(aStruct).WhereStruct(aStruct, includeZeroValue...)
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

func (b *QueryBuilder) Columns(columns ...string) *QueryBuilder {
	for _, column := range columns {
		b.columns = append(b.columns, NewTemplate(column))
	}
	return b
}

func (b *QueryBuilder) TemplateColumns(columns ...Template) *QueryBuilder {
	b.columns = append(b.columns, columns...)
	return b
}

func (b *QueryBuilder) StructColumns(aStruct interface{}) *QueryBuilder {
	names := structFields(reflect.TypeOf(aStruct)).columnNames()
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

func (b *QueryBuilder) Join(format string, values ...interface{}) *QueryBuilder {
	b.joins = append(b.joins, Template{Format: format, Values: values})
	return b
}

func (b *QueryBuilder) JoinTemplate(templates ...Template) *QueryBuilder {
	b.joins = append(b.joins, templates...)
	return b
}

func (b *QueryBuilder) Where(format string, values ...interface{}) *QueryBuilder {
	b.where = b.where.AppendTemplate(NewTemplate(format, values...))
	return b
}

func (b *QueryBuilder) WhereTemplate(templates ...Template) *QueryBuilder {
	b.where = b.where.AppendTemplate(templates...)
	return b
}

func (b *QueryBuilder) WhereMap(m map[string]interface{}) *QueryBuilder {
	b.where = b.where.AppendMap(m)
	return b
}

func (b *QueryBuilder) WhereStruct(i interface{}, includeZeroValue ...bool) *QueryBuilder {
	b.where = b.where.AppendStruct(i, includeZeroValue...)
	return b
}

func (b *QueryBuilder) Having(format string, values ...interface{}) *QueryBuilder {
	b.having.AppendTemplate(Template{Format: format, Values: values})
	return b
}

func (b *QueryBuilder) HavingTemplate(templates ...Template) *QueryBuilder {
	b.having.AppendTemplate(templates...)
	return b
}

func (b *QueryBuilder) HavingMap(m map[string]interface{}) *QueryBuilder {
	b.where.AppendMap(m)
	return b
}

func (b *QueryBuilder) HavingStruct(i interface{}) *QueryBuilder {
	b.where.AppendStruct(i)
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

func (b *QueryBuilder) Paging(format string, values ...interface{}) *QueryBuilder {
	b.paging = NewTemplate(format, values...)
	return b
}

func (b *QueryBuilder) PagingTemplate(template Template) *QueryBuilder {
	b.paging = template
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

func (b *QueryBuilder) Query(db DB) (*Rows, error) {
	return b.Build().Query(db)
}

func (b *QueryBuilder) QueryContext(ctx context.Context, db DB) (*Rows, error) {
	return b.Build().QueryContext(ctx, db)
}

func (b *QueryBuilder) QueryScalar(ctx context.Context, db DB, value interface{}) error {
	return b.Build().QueryScalar(ctx, db, value)
}

func (b *QueryBuilder) QueryScalarSlice(ctx context.Context, db DB, values interface{}) error {
	return b.Build().QueryScalarSlice(ctx, db, values)
}

func (b *QueryBuilder) QueryMap(ctx context.Context, db DB) (map[string]interface{}, error) {
	return b.Build().QueryMap(ctx, db)
}

func (b *QueryBuilder) QueryMapSlice(ctx context.Context, db DB) ([]map[string]interface{}, error) {
	return b.Build().QueryMapSlice(ctx, db)
}

func (b *QueryBuilder) QueryStruct(ctx context.Context, db DB, structPtr interface{}) error {
	return b.Build().QueryStruct(ctx, db, structPtr)
}

func (b *QueryBuilder) QueryStructSlice(ctx context.Context, db DB, structPtr interface{}) error {
	return b.Build().QueryStructSlice(ctx, db, structPtr)
}

func (b *QueryBuilder) QuerySelectStruct(ctx context.Context, db DB, structPtr interface{}) error {
	b.StructColumns(structPtr)
	return b.QueryStruct(ctx, db, structPtr)
}

func (b *QueryBuilder) QuerySelectStructSlice(ctx context.Context, db DB, structPtr interface{}) error {
	typo := reflect.TypeOf(structPtr)
	for typo.Kind() == reflect.Ptr {
		typo = typo.Elem()
	}
	if typo.Kind() != reflect.Slice {
		return errors.New("bear: require struct slice")
	}
	b.StructColumns(reflect.New(typo.Elem()).Elem().Interface())
	return b.QueryStructSlice(ctx, db, structPtr)
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

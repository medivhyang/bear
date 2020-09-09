package bear

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type (
	Executor interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
	}
	WithContextExectutor interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}
	Querier interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}
	WithContextQuerier interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}
)

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
	if !where.IsEmpty() {
		result = result.Join(where, " where ")
	}

	if len(b.groupBy) > 0 {
		result = result.Append(fmt.Sprintf(" group by %s", strings.Join(b.groupBy, ",")))
	}

	having := b.having.Build()
	if !having.IsEmpty() {
		result = result.Join(having, " having ")
	}

	if len(b.orderBy) > 0 {
		result = result.Append(fmt.Sprintf(" order by %s", strings.Join(b.orderBy, ",")))
	}

	if !b.paging.IsEmpty() {
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

const (
	actionInsert = "insert"
	actionUpdate = "update"
	actionDelete = "delete"
)

func Insert(table string, pairs map[string]interface{}) *commandBuilder {
	b := &commandBuilder{table: table, action: actionInsert}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func InsertStruct(i interface{}, ignoreColumns ...string) *commandBuilder {
	return Insert(TableName(i), structToColumnValueMap(reflect.ValueOf(i), true, ignoreColumns...))
}

func Update(table string, pairs map[string]interface{}) *commandBuilder {
	b := &commandBuilder{table: table, action: actionUpdate}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func UpdateStruct(i interface{}) *commandBuilder {
	return Update(TableName(i), structToColumnValueMap(reflect.ValueOf(i), false))
}

func Delete(table string) *commandBuilder {
	b := &commandBuilder{table: table, action: actionDelete}
	return b
}

type commandBuilder struct {
	action  string
	dialect string
	table   string
	names   []string
	values  []interface{}
	where   Condition
}

func (b *commandBuilder) Dialect(name string) *commandBuilder {
	b.dialect = name
	return b
}

func (b *commandBuilder) Table(name string) *commandBuilder {
	b.table = name
	return b
}

func (b *commandBuilder) Where(format string, values ...interface{}) *commandBuilder {
	b.where = b.where.AppendTemplate(Template{Format: format, Values: values})
	return b
}

func (b *commandBuilder) WhereTemplate(templates ...Template) *commandBuilder {
	b.where = b.where.AppendTemplate(templates...)
	return b
}

func (b *commandBuilder) WhereMap(m map[string]interface{}) *commandBuilder {
	b.where = b.where.AppendMap(m)
	return b
}

func (b *commandBuilder) WhereStruct(i interface{}) *commandBuilder {
	b.where = b.where.AppendStruct(i)
	return b
}

func (b *commandBuilder) Build() Template {
	result := Template{}

	switch b.action {
	case actionInsert:
		var holders []string
		for i := 0; i < len(b.values); i++ {
			holders = append(holders, "?")
		}
		result.Format = fmt.Sprintf("insert into %s(%s) values(%s)",
			b.table,
			strings.Join(b.names, ","),
			strings.Join(holders, ","),
		)
		result.Values = append(result.Values, b.values...)
	case actionUpdate:
		var pairs []string
		for i := 0; i < len(b.names); i++ {
			pairs = append(pairs, fmt.Sprintf("%s=?", b.names[i]))
		}
		result.Format = fmt.Sprintf("update %s set %s",
			b.table,
			strings.Join(pairs, ","),
		)
		result.Values = append(result.Values, b.values...)
	case actionDelete:
		result.Format = fmt.Sprintf("delete from %s", b.table)
	}

	where := b.where.Build()
	if !where.IsEmpty() {
		result = result.Join(where, " where ")
	}

	return result
}

func (b *commandBuilder) Execute(exectutor Executor) (*Result, error) {
	return b.Build().Execute(exectutor)
}

func (b *commandBuilder) ExecuteContext(ctx context.Context, exectutor WithContextExectutor) (*Result, error) {
	return b.Build().ExecuteContext(ctx, exectutor)
}

const (
	actionCreateTable                      = "create_table"
	actionCreateTableIfNotExists           = "create_table_if_not_exists"
	actionCreateTableWithStruct            = "create_table_with_struct"
	actionCreateTableWithStructIfNotExists = "create_table_with_struct_if_not_exists"
	actionDropTable                        = "drop_table"
	actionDropTableIfExists                = "drop_table_if_exists"
)

type Table struct {
	Table   string
	Columns []Column
}

type Column struct {
	Name   string
	Type   string
	Suffix string
}

func CreateTable(table string, columns []Column) *tableBuilder {
	return &tableBuilder{
		action:  actionCreateTable,
		table:   table,
		columns: columns,
	}
}

func BatchCreateTable(tables []Table) Template {
	var result Template
	for _, table := range tables {
		result.Join(CreateTable(table.Table, table.Columns).Build(), "\n")
	}
	return result
}

func CreateTableStruct(i interface{}) *tableBuilder {
	return &tableBuilder{
		action:     actionCreateTableWithStruct,
		table:      TableName(i),
		structType: reflect.TypeOf(i),
	}
}

func CreateTableIfNotExists(table string, columns []Column) *tableBuilder {
	return &tableBuilder{
		action:  actionCreateTableIfNotExists,
		table:   table,
		columns: columns,
	}
}

func CreateTableIfNotExistsStruct(i interface{}) *tableBuilder {
	return &tableBuilder{
		action:     actionCreateTableWithStructIfNotExists,
		table:      TableName(i),
		structType: reflect.TypeOf(i),
	}
}

func DropTable(table string) *tableBuilder {
	return &tableBuilder{
		action: actionDropTable,
		table:  table,
	}
}

func DropTableIfExists(table string) *tableBuilder {
	return &tableBuilder{
		action: actionDropTableIfExists,
		table:  table,
	}
}

func DropTableStruct(i interface{}) *tableBuilder {
	return &tableBuilder{
		action: actionDropTable,
		table:  TableName(i),
	}
}

func DropTableIfExistsStruct(i interface{}) *tableBuilder {
	return &tableBuilder{
		action: actionDropTableIfExists,
		table:  TableName(i),
	}
}

func (c Column) Build() Template {
	template := fmt.Sprintf("%s %s", c.Name, c.Type)
	suffix := strings.TrimSpace(c.Suffix)
	if suffix != "" {
		template += " " + suffix
	}
	return NewTemplate(template)
}

type tableBuilder struct {
	action     string
	dialect    string
	table      string
	columns    []Column
	structType reflect.Type
	prepends   []Template
	appends    []Template
}

func (b *tableBuilder) Dialect(name string) *tableBuilder {
	b.dialect = name
	return b
}

func (b *tableBuilder) Table(name string) *tableBuilder {
	b.table = name
	return b
}

func (b *tableBuilder) Prepend(template string, values ...interface{}) *tableBuilder {
	b.prepends = append(b.prepends, NewTemplate(template, values...))
	return b
}

func (b *tableBuilder) Append(template string, values ...interface{}) *tableBuilder {
	b.appends = append(b.appends, NewTemplate(template, values...))
	return b
}

func (b *tableBuilder) Build() Template {
	result := Template{}
	buffer := strings.Builder{}

	for _, item := range b.prepends {
		buffer.WriteString(item.Format)
		buffer.WriteString("\n")
		result.Values = append(result.Values, item.Values...)
	}

	switch b.action {
	case actionCreateTable, actionCreateTableIfNotExists,
		actionCreateTableWithStruct, actionCreateTableWithStructIfNotExists:
		if b.action == actionCreateTableWithStruct || b.action == actionCreateTableWithStructIfNotExists {
			b.columns = structFields(b.structType).columns(b.dialect)
		}
		if len(b.columns) == 0 {
			break
		}
		if buffer.Len() > 0 && !strings.HasSuffix(buffer.String(), "\n") {
			buffer.WriteString("\n")
		}
		switch b.action {
		case actionCreateTable, actionCreateTableWithStruct:
			buffer.WriteString(fmt.Sprintf("create table %s (\n", b.table))
		case actionCreateTableIfNotExists, actionCreateTableWithStructIfNotExists:
			buffer.WriteString(fmt.Sprintf("create table if not exists %s (\n", b.table))
		}
		for i, c := range b.columns {
			s := c.Build()
			buffer.WriteString("  ")
			buffer.WriteString(s.Format)
			if i < len(b.columns)-1 {
				buffer.WriteString(",\n")
			} else {
				buffer.WriteString("\n")
			}
			result.Values = append(result.Values, s.Values...)
		}
		buffer.WriteString(");\n")
	case actionDropTable:
		buffer.WriteString(fmt.Sprintf("drop table %s;\n", b.table))
	case actionDropTableIfExists:
		buffer.WriteString(fmt.Sprintf("drop table if exists %s;\n", b.table))
	}

	for _, item := range b.appends {
		buffer.WriteString(item.Format)
		buffer.WriteString("\n")
		result.Values = append(result.Values, item.Values...)
	}

	result.Format = buffer.String()

	return result
}

func (b *tableBuilder) Execute(executor Executor) (*Result, error) {
	return b.Build().Execute(executor)
}

func (b *tableBuilder) ExecuteContext(ctx context.Context, executor WithContextExectutor) (*Result, error) {
	return b.Build().ExecuteContext(ctx, executor)
}

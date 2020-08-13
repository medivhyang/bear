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

func Select(table string, names ...string) *queryBuilder {
	b := queryBuilder{table: table}
	for _, name := range names {
		b.fields = append(b.fields, Template{Format: name})
	}
	return &b
}

func SelectWithTemplate(table string, templates ...Template) *queryBuilder {
	b := queryBuilder{table: table}
	b.fields = append(b.fields, templates...)
	return &b
}

func SelectWithStruct(i interface{}) *queryBuilder {
	return Select(TableName(i), structFields(reflect.TypeOf(i)).dbFieldNames()...)
}

func SelectWhere(i interface{}) *queryBuilder {
	return Select(TableName(i), structFields(reflect.TypeOf(i)).dbFieldNames()...).WhereWithStruct(i)
}

type queryBuilder struct {
	dialect string
	table   string
	fields  []Template
	joins   []Template
	where   conditionBuilder
	groupBy []string
	having  conditionBuilder
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

func (b *queryBuilder) JoinWithTemplate(templates ...Template) *queryBuilder {
	b.joins = append(b.joins, templates...)
	return b
}

func (b *queryBuilder) Where(format string, values ...interface{}) *queryBuilder {
	b.where.Append(Template{Format: format, Values: values})
	return b
}

func (b *queryBuilder) WhereWithTemplate(templates ...Template) *queryBuilder {
	b.where.Append(templates...)
	return b
}

func (b *queryBuilder) WhereWithMap(m map[string]interface{}) *queryBuilder {
	b.where.WithMap(m)
	return b
}

func (b *queryBuilder) WhereWithStruct(i interface{}) *queryBuilder {
	b.where.WithStruct(i)
	return b
}

func (b *queryBuilder) Having(format string, values ...interface{}) *queryBuilder {
	b.having.Append(Template{Format: format, Values: values})
	return b
}

func (b *queryBuilder) HavingWithTemplate(templates ...Template) *queryBuilder {
	b.having.Append(templates...)
	return b
}

func (b *queryBuilder) HavingWithMap(m map[string]interface{}) *queryBuilder {
	b.where.WithMap(m)
	return b
}

func (b *queryBuilder) HavingWithStruct(i interface{}) *queryBuilder {
	b.where.WithStruct(i)
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

func (b *queryBuilder) PagingWithLimit(offset int, limit int) *queryBuilder {
	b.paging = New("limit ?,?", offset, limit)
	return b
}

func (b *queryBuilder) Paging(format string, values ...interface{}) *queryBuilder {
	b.paging = New(format, values...)
	return b
}

func (b *queryBuilder) PagingWithTemplate(template Template) *queryBuilder {
	b.paging = template
	return b
}

func (b *queryBuilder) Build() Template {
	result := Template{}

	var fieldsFormats []string
	for _, field := range b.fields {
		fieldsFormats = append(fieldsFormats, field.Format)
		result.Values = append(result.Values, field.Values...)
	}
	result.Format = fmt.Sprintf("select %s from %s",
		strings.Join(fieldsFormats, ","),
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

func (b *queryBuilder) QueryWithContext(querier WithContextQuerier, ctx context.Context) (*Rows, error) {
	return b.Build().QueryWithContext(querier, ctx)
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

func InsertWithStruct(i interface{}) *commandBuilder {
	return Insert(TableName(i), structToValueMap(reflect.ValueOf(i), true))
}

func Update(table string, pairs map[string]interface{}) *commandBuilder {
	b := &commandBuilder{table: table, action: actionUpdate}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func UpdateWithStruct(i interface{}) *commandBuilder {
	return Update(TableName(i), structToValueMap(reflect.ValueOf(i), false))
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
	where   conditionBuilder
}

func (b *commandBuilder) Dialect(name string) *commandBuilder {
	b.dialect = name
	return b
}

func (b *commandBuilder) Table(name string) *commandBuilder {
	b.table = name
	return b
}

func (b *commandBuilder) Where(template string, values ...interface{}) *commandBuilder {
	b.where.Append(Template{Format: template, Values: values})
	return b
}

func (b *commandBuilder) WhereWithTemplate(templates ...Template) *commandBuilder {
	b.where.Append(templates...)
	return b
}

func (b *commandBuilder) WhereWithMap(m map[string]interface{}) *commandBuilder {
	b.where.WithMap(m)
	return b
}

func (b *commandBuilder) WhereWithStruct(i interface{}) *commandBuilder {
	b.where.WithStruct(i)
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

func (b *commandBuilder) ExecuteWithContext(exectutor WithContextExectutor, ctx context.Context) (*Result, error) {
	return b.Build().ExecuteWitchContext(exectutor, ctx)
}

type conditionBuilder struct {
	conditions []string
	names      []string
	values     []interface{}
}

func (b *conditionBuilder) Append(items ...Template) *conditionBuilder {
	for _, item := range items {
		b.conditions = append(b.conditions, item.Format)
		b.values = append(b.values, item.Values...)
	}
	return b
}

func (b *conditionBuilder) WithMap(m map[string]interface{}) *conditionBuilder {
	var (
		items  []string
		values []interface{}
	)
	for k, v := range m {
		items = append(items, fmt.Sprintf("%s = ?", k))
		values = append(values, v)
	}
	template := strings.Join(items, " and ")
	if template != "" {
		b.conditions = append(b.conditions, template)
		b.values = append(b.values, values...)
	}
	return b
}

func (b *conditionBuilder) WithStruct(i interface{}) *conditionBuilder {
	var (
		items  []string
		values []interface{}
	)
	m := structToValueMap(reflect.ValueOf(i), false)
	for k, v := range m {
		items = append(items, fmt.Sprintf("%s = ?", k))
		values = append(values, v)
	}
	template := strings.Join(items, " and ")
	if template != "" {
		b.conditions = append(b.conditions, template)
		b.values = append(b.values, values...)
	}
	return b
}

func (b *conditionBuilder) Build() Template {
	if len(b.conditions) == 0 {
		return Template{}
	}

	names := make([]string, len(b.names))
	copy(names, b.names)

	values := make([]interface{}, len(b.values))
	copy(values, b.values)

	var conditions []string
	for _, condition := range b.conditions {
		conditions = append(conditions, "("+condition+")")
	}

	result := Template{
		Format: strings.Join(conditions, " and "),
		Values: values,
	}

	return result
}

const (
	actionCreateTable                      = "create_table"
	actionCreateTableIfNotExists           = "create_table_if_not_exists"
	actionCreateTableWithStruct            = "create_table_with_struct"
	actionCreateTableWithStructIfNotExists = "create_table_with_struct_if_not_exists"
	actionDropTable                        = "drop_table"
	actionDropTableIfExists                = "drop_table_if_exists"
)

type DBField struct {
	Name   string
	Type   string
	Suffix string
}

func CreateTable(table string, fields []DBField) *TableBuilder {
	return &TableBuilder{
		action: actionCreateTable,
		table:  table,
		fields: fields,
	}
}

func CreateTableWithStruct(i interface{}) *TableBuilder {
	return &TableBuilder{
		action:     actionCreateTableWithStruct,
		table:      TableName(i),
		structType: reflect.TypeOf(i),
	}
}

func CreateTableWithStructIfNotExists(i interface{}) *TableBuilder {
	return &TableBuilder{
		action:     actionCreateTableWithStructIfNotExists,
		table:      TableName(i),
		structType: reflect.TypeOf(i),
	}
}

func CreateTableIfNotExists(table string, fields []DBField) *TableBuilder {
	return &TableBuilder{
		action: actionCreateTableIfNotExists,
		table:  table,
		fields: fields,
	}
}

func DropTable(table string) *TableBuilder {
	return &TableBuilder{
		action: actionDropTable,
		table:  table,
	}
}

func DropTableIfExists(table string) *TableBuilder {
	return &TableBuilder{
		action: actionDropTableIfExists,
		table:  table,
	}
}

func DropTableWithStruct(i interface{}) *TableBuilder {
	return &TableBuilder{
		action: actionDropTable,
		table:  TableName(i),
	}
}

func DropTableWithStructIfExists(i interface{}) *TableBuilder {
	return &TableBuilder{
		action: actionDropTableIfExists,
		table:  TableName(i),
	}
}

func (field DBField) Build() Template {
	template := fmt.Sprintf("%s %s", field.Name, field.Type)
	if field.Suffix != "" {
		template += field.Suffix
	}
	return New(template)
}

type dbFieldSlice []DBField

func (fields dbFieldSlice) names() []string {
	var result []string
	for _, field := range fields {
		result = append(result, field.Name)
	}
	return result
}

type TableBuilder struct {
	action     string
	dialet     string
	table      string
	fields     dbFieldSlice
	structType reflect.Type
	prepends   []Template
	appends    []Template
}

func (b *TableBuilder) Dialect(name string) *TableBuilder {
	b.dialet = name
	return b
}

func (b *TableBuilder) Table(name string) *TableBuilder {
	b.table = name
	return b
}

func (b *TableBuilder) Prepend(template string, values ...interface{}) *TableBuilder {
	b.prepends = append(b.prepends, New(template, values...))
	return b
}

func (b *TableBuilder) Append(template string, values ...interface{}) *TableBuilder {
	b.appends = append(b.appends, New(template, values...))
	return b
}

func (b *TableBuilder) Build() Template {
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
			b.fields = structFields(b.structType).dbFields(b.dialet)
		}
		if len(b.fields) == 0 {
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
		for i, field := range b.fields {
			s := field.Build()
			buffer.WriteString("  ")
			buffer.WriteString(s.Format)
			if i < len(b.fields)-1 {
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

func (b *TableBuilder) Execute(executor Executor) (*Result, error) {
	return b.Build().Execute(executor)
}

func (b *TableBuilder) ExecuteWitchContext(executor WithContextExectutor, ctx context.Context) (*Result, error) {
	return b.Build().ExecuteWitchContext(executor, ctx)
}

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

func Select(table string, fields ...Template) *QueryBuilder {
	b := QueryBuilder{table: table}
	b.fields = append(b.fields, fields...)
	return &b
}

func SelectSimple(table string, names ...string) *QueryBuilder {
	b := QueryBuilder{table: table}
	for _, name := range names {
		b.fields = append(b.fields, Template{Format: name})
	}
	return &b
}

func SelectWithStruct(i interface{}) *QueryBuilder {
	return SelectSimple(TableName(i), structFields(reflect.TypeOf(i)).dbFieldNames()...)
}

func SelectWhere(i interface{}) *QueryBuilder {
	return SelectSimple(TableName(i), structFields(reflect.TypeOf(i)).dbFieldNames()...).WhereWithStruct(i)
}

type QueryBuilder struct {
	dialect string
	table   string
	fields  []Template
	joins   []Template
	where   ConditionBuilder
	groupBy []string
	having  ConditionBuilder
	orderBy []string
	paging  Template
}

func (b *QueryBuilder) Dialect(name string) *QueryBuilder {
	b.dialect = name
	return b
}

func (b *QueryBuilder) Table(name string) *QueryBuilder {
	b.table = name
	return b
}

func (b *QueryBuilder) Join(format string, values ...interface{}) *QueryBuilder {
	b.joins = append(b.joins, Template{Format: format, Values: values})
	return b
}

func (b *QueryBuilder) JoinWithTemplate(templates ...Template) *QueryBuilder {
	b.joins = append(b.joins, templates...)
	return b
}

func (b *QueryBuilder) Where(format string, values ...interface{}) *QueryBuilder {
	b.where.Append(Template{Format: format, Values: values})
	return b
}

func (b *QueryBuilder) WhereWithTemplate(templates ...Template) *QueryBuilder {
	b.where.Append(templates...)
	return b
}

func (b *QueryBuilder) WhereWithMap(m map[string]interface{}) *QueryBuilder {
	b.where.WithMap(m)
	return b
}

func (b *QueryBuilder) WhereWithStruct(i interface{}) *QueryBuilder {
	b.where.WithStruct(i)
	return b
}

func (b *QueryBuilder) Having(format string, values ...interface{}) *QueryBuilder {
	b.having.Append(Template{Format: format, Values: values})
	return b
}

func (b *QueryBuilder) HavingWithTemplate(templates ...Template) *QueryBuilder {
	b.having.Append(templates...)
	return b
}

func (b *QueryBuilder) HavingWithMap(m map[string]interface{}) *QueryBuilder {
	b.where.WithMap(m)
	return b
}

func (b *QueryBuilder) HavingWithStruct(i interface{}) *QueryBuilder {
	b.where.WithStruct(i)
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

func (b *QueryBuilder) PagingWithLimit(offset int, limit int) *QueryBuilder {
	b.paging = New("limit ?,?", offset, limit)
	return b
}

func (b *QueryBuilder) Paging(format string, values ...interface{}) *QueryBuilder {
	b.paging = New(format, values...)
	return b
}

func (b *QueryBuilder) PagingWithTemplate(template Template) *QueryBuilder {
	b.paging = template
	return b
}

func (b *QueryBuilder) Build() Template {
	result := Template{}

	var fieldsTemplates []string
	for _, field := range b.fields {
		fieldsTemplates = append(fieldsTemplates, field.Format)
		result.Values = append(result.Values, field.Values...)
	}
	result.Format = fmt.Sprintf("select %s from %s",
		strings.Join(fieldsTemplates, ","),
		b.table,
	)

	for _, item := range b.joins {
		result = result.Append(" "+item.Format, item.Values...)
	}

	where := b.where.Build()
	if len(where.Format) > 0 {
		result.Format += " where " + where.Format
		result.Values = append(result.Values, where.Values...)
	}

	if len(b.groupBy) > 0 {
		result.Format += fmt.Sprintf(" group by %s", strings.Join(b.groupBy, ","))
	}

	having := b.having.Build()
	if len(having.Format) > 0 {
		result.Format += " having" + having.Format
		result.Values = append(result.Values, having.Values...)
	}

	if len(b.orderBy) > 0 {
		result.Format += fmt.Sprintf(" order by %s", strings.Join(b.orderBy, ","))
	}

	if !b.paging.IsEmpty() {
		result = result.Join(b.paging, " ")
	}

	return result
}

func (b *QueryBuilder) As(alias string) Template {
	s := b.Build()
	s.Format = fmt.Sprintf("(%s) as %s", s.Format, alias)
	return s
}

func (b *QueryBuilder) Query(querier Querier) (*Rows, error) {
	return b.Build().Query(querier)
}

func (b *QueryBuilder) QueryWithContext(querier WithContextQuerier, ctx context.Context) (*Rows, error) {
	return b.Build().QueryWithContext(querier, ctx)
}

const (
	actionInsert = "insert"
	actionUpdate = "update"
	actionDelete = "delete"
)

func Insert(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: actionInsert}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func InsertWithStruct(i interface{}) *CommandBuilder {
	return Insert(TableName(i), structToMap(reflect.ValueOf(i)))
}

func Update(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: actionUpdate}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func UpdateWithStruct(i interface{}) *CommandBuilder {
	return Update(TableName(i), structToMap(reflect.ValueOf(i)))
}

func Delete(table string) *CommandBuilder {
	b := &CommandBuilder{table: table, action: actionDelete}
	return b
}

type CommandBuilder struct {
	action  string
	dialect string
	table   string
	names   []string
	values  []interface{}
	where   ConditionBuilder
}

func (b *CommandBuilder) Dialect(name string) *CommandBuilder {
	b.dialect = name
	return b
}

func (b *CommandBuilder) Table(name string) *CommandBuilder {
	b.table = name
	return b
}

func (b *CommandBuilder) Where(template string, values ...interface{}) *CommandBuilder {
	b.where.Append(Template{Format: template, Values: values})
	return b
}

func (b *CommandBuilder) WhereWithTemplate(templates ...Template) *CommandBuilder {
	b.where.Append(templates...)
	return b
}

func (b *CommandBuilder) WhereWithMap(m map[string]interface{}) *CommandBuilder {
	b.where.WithMap(m)
	return b
}

func (b *CommandBuilder) WhereWithStruct(i interface{}) *CommandBuilder {
	b.where.WithStruct(i)
	return b
}

func (b *CommandBuilder) Build() Template {
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
	if len(where.Format) > 0 {
		result.Format += " where " + where.Format
		result.Values = append(result.Values, where.Values...)
	}

	return result
}

func (b *CommandBuilder) Execute(exectutor Executor) (*Result, error) {
	return b.Build().Execute(exectutor)
}

func (b *CommandBuilder) ExecuteWithContext(exectutor WithContextExectutor, ctx context.Context) (*Result, error) {
	return b.Build().ExecuteWitchContext(exectutor, ctx)
}

type ConditionBuilder struct {
	conditions []string
	names      []string
	values     []interface{}
}

func (b *ConditionBuilder) Append(items ...Template) *ConditionBuilder {
	for _, item := range items {
		b.conditions = append(b.conditions, item.Format)
		b.values = append(b.values, item.Values...)
	}
	return b
}

func (b *ConditionBuilder) WithMap(m map[string]interface{}) *ConditionBuilder {
	var (
		items  []string
		values []interface{}
	)
	for k, v := range m {
		if !isZeroValue(reflect.ValueOf(v)) {
			items = append(items, fmt.Sprintf("%s = ?", k))
			values = append(values, v)
		}
	}
	template := strings.Join(items, " and ")
	if template != "" {
		b.conditions = append(b.conditions, template)
		b.values = append(b.values, values...)
	}
	return b
}

func (b *ConditionBuilder) WithStruct(i interface{}) *ConditionBuilder {
	var (
		items  []string
		values []interface{}
	)
	m := structToMap(reflect.ValueOf(i))
	for k, v := range m {
		if !isZeroValue(reflect.ValueOf(v)) {
			items = append(items, fmt.Sprintf("%s = ?", k))
			values = append(values, v)
		}
	}
	template := strings.Join(items, " and ")
	if template != "" {
		b.conditions = append(b.conditions, template)
		b.values = append(b.values, values...)
	}
	return b
}

func (b *ConditionBuilder) Build() Template {
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

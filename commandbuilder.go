package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

const (
	actionInsert = "insert"
	actionUpdate = "update"
	actionDelete = "delete"
)

type commandBuilder struct {
	action  string
	dialect string
	table   string
	names   []string
	values  []interface{}
	where   Condition
}

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
	if !where.IsEmptyOrWhitespace() {
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

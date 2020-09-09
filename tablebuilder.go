package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

const (
	actionCreateTable                      = "create_table"
	actionCreateTableIfNotExists           = "create_table_if_not_exists"
	actionCreateTableWithStruct            = "create_table_with_struct"
	actionCreateTableWithStructIfNotExists = "create_table_with_struct_if_not_exists"
	actionDropTable                        = "drop_table"
	actionDropTableIfExists                = "drop_table_if_exists"
)

type tableBuilder struct {
	action     string
	dialect    string
	table      string
	columns    []Column
	structType reflect.Type
	prepends   []Template
	appends    []Template
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

type Table struct {
	Table   string
	Columns []Column
}

type Column struct {
	Name   string
	Type   string
	Suffix string
}

func (c Column) Build() Template {
	format := fmt.Sprintf("%s %s", c.Name, c.Type)
	suffix := strings.TrimSpace(c.Suffix)
	if suffix != "" {
		format += " " + suffix
	}
	return NewTemplate(format)
}

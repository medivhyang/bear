package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

const (
	actionCreateTable = "create_table"
	actionDropTable   = "drop_table"
)

type tableBuilder struct {
	action        string
	dialect       string
	table         string
	columns       []Column
	include       map[string]bool
	exclude       map[string]bool
	prepends      []Template
	appends       []Template
	innerPrepends []Template
	innerAppends  []Template
	newline       bool
	indent        string
	prefix        string
	onExists      bool
	onNotExists   bool
}

func CreateTable(table string, columns []Column) *tableBuilder {
	return &tableBuilder{
		action:  actionCreateTable,
		table:   table,
		columns: columns,
	}
}

func CreateTableStruct(aStruct interface{}, dialect ...string) *tableBuilder {
	finalDialect := ""
	if len(dialect) > 0 {
		finalDialect = dialect[0]
	}
	return &tableBuilder{
		action:  actionCreateTable,
		table:   TableName(aStruct),
		columns: structFields(reflect.TypeOf(aStruct)).columns(finalDialect),
	}
}

func BatchCreateTable(tables []Table) Template {
	var result Template
	for _, table := range tables {
		result = result.Join(CreateTable(table.Name, table.Columns).Build(), "\n")
	}
	return result
}

func DropTable(table string) *tableBuilder {
	return &tableBuilder{
		action: actionDropTable,
		table:  table,
	}
}

func DropTableStruct(aStruct interface{}) *tableBuilder {
	return &tableBuilder{
		action: actionDropTable,
		table:  TableName(aStruct),
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

func (b *tableBuilder) OnExists() *tableBuilder {
	b.onExists = true
	return b
}

func (b *tableBuilder) OnNotExists() *tableBuilder {
	b.onNotExists = true
	return b
}

func (b *tableBuilder) Indent(prefix string, indent string) *tableBuilder {
	b.newline = true
	b.prefix = prefix
	b.indent = indent
	return b
}

func (b *tableBuilder) Include(names ...string) *tableBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *tableBuilder) Exclude(names ...string) *tableBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
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

func (b *tableBuilder) PrependInner(template string, values ...interface{}) *tableBuilder {
	b.innerPrepends = append(b.innerPrepends, NewTemplate(template, values...))
	return b
}

func (b *tableBuilder) AppendInner(template string, values ...interface{}) *tableBuilder {
	b.innerAppends = append(b.innerAppends, NewTemplate(template, values...))
	return b
}

func (b *tableBuilder) Build() Template {
	result := Template{}
	buffer := strings.Builder{}

	for _, item := range b.prepends {
		buffer.WriteString(item.Format)
		if b.newline {
			buffer.WriteString("\n")
		}
		result.Values = append(result.Values, item.Values...)
	}

	switch b.action {
	case actionCreateTable:
		var columns = b.finalColumns()
		if len(columns) == 0 {
			break
		}
		if b.newline {
			if buffer.Len() > 0 && !strings.HasSuffix(buffer.String(), "\n") {
				buffer.WriteString("\n")
			}
		}
		if b.onNotExists {
			buffer.WriteString(fmt.Sprintf("create table if not exists %s (", b.table))
		} else {
			buffer.WriteString(fmt.Sprintf("create table %s (", b.table))
		}
		if b.newline {
			buffer.WriteString("\n")
		}
		for _, item := range b.innerPrepends {
			buffer.WriteString(b.indent)
			buffer.WriteString(b.prefix)
			buffer.WriteString(item.Format)
			if b.newline {
				buffer.WriteString("\n")
			}
			result.Values = append(result.Values, item.Values...)
		}
		for _, column := range columns {
			buffer.WriteString(b.indent)
			buffer.WriteString(b.prefix)
			buffer.WriteString(fmt.Sprintf("%s %s", column.Name, column.Type))
			suffix := strings.TrimSpace(column.Suffix)
			if suffix != "" {
				buffer.WriteString(" ")
				buffer.WriteString(suffix)
			}
			buffer.WriteString(",")
			if b.newline {
				buffer.WriteString("\n")
			}
		}
		for _, item := range b.innerAppends {
			buffer.WriteString(b.indent)
			buffer.WriteString(b.prefix)
			buffer.WriteString(item.Format)
			if b.newline {
				buffer.WriteString("\n")
			}
			result.Values = append(result.Values, item.Values...)
		}
		newBufferContent := strings.TrimRight(buffer.String(), ", \n")
		buffer.Reset()
		buffer.WriteString(newBufferContent)
		if b.newline {
			buffer.WriteString("\n")
		}
		buffer.WriteString(");")
		if b.newline {
			buffer.WriteString("\n")
		}
	case actionDropTable:
		if b.onExists {
			buffer.WriteString(fmt.Sprintf("drop table if exists %s;", b.table))
		} else {
			buffer.WriteString(fmt.Sprintf("drop table %s;", b.table))
		}
	}
	for _, item := range b.appends {
		buffer.WriteString(item.Format)
		if b.newline {
			buffer.WriteString("\n")
		}
		result.Values = append(result.Values, item.Values...)
	}
	result.Format = strings.TrimSpace(buffer.String())
	if b.newline && !strings.HasSuffix(result.Format, "\n") {
		result.Format += "\n"
	}
	return result
}

func (b *tableBuilder) Execute(executor Executor) (*Result, error) {
	return b.Build().Execute(executor)
}

func (b *tableBuilder) ExecuteContext(ctx context.Context, executor WithContextExectutor) (*Result, error) {
	return b.Build().ExecuteContext(ctx, executor)
}

func (b *tableBuilder) finalColumns() []Column {
	var includedColumns []Column
	if len(b.include) > 0 {
		for _, column := range b.columns {
			if b.include[column.Name] {
				includedColumns = append(includedColumns, column)
			}
		}
	} else {
		includedColumns = b.columns
	}
	var excludedColumns []Column
	if len(b.exclude) > 0 {
		for _, column := range includedColumns {
			if !b.exclude[column.Name] {
				excludedColumns = append(excludedColumns, column)
			}
		}
	} else {
		excludedColumns = includedColumns
	}
	return excludedColumns
}

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name   string
	Type   string
	Suffix string
}

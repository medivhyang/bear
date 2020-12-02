package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

type ColumnSchema struct {
	Name   string
	Type   string
	Suffix string
}

type TableAction string

const (
	TableActionCreateTable TableAction = "create_table"
	TableActionDropTable   TableAction = "drop_table"
)

func (a TableAction) Valid() bool {
	switch a {
	case TableActionCreateTable, TableActionDropTable:
		return true
	default:
		return false
	}
}

type TableBuilder struct {
	action           TableAction
	dialect          string
	table            string
	columns          []ColumnSchema
	include          map[string]bool
	exclude          map[string]bool
	prepends         []Template
	appends          []Template
	innerPrepends    []Template
	innerAppends     []Template
	newline          bool
	indent           string
	prefix           string
	onlyDropIfExists bool
	onCreateIfExists bool
}

func NewTableBuilder(dialect ...string) *TableBuilder {
	b := &TableBuilder{}
	if len(dialect) > 0 {
		b.Dialect(dialect[0])
	}
	return b
}

func CreateTable(table string, columns ...interface{}) *TableBuilder {
	return NewTableBuilder().CreateTable(table, columns...)
}

func CreateTableIfNotExists(table string, columns ...interface{}) *TableBuilder {
	return NewTableBuilder().CreateTableIfNotExists(table, columns...)
}

func DropTable(table string) *TableBuilder {
	return NewTableBuilder().DropTable(table)
}

func DropTableIfExists(table string) *TableBuilder {
	return NewTableBuilder().DropTableIfExists(table)
}

func BatchCreateTables(dialect string, args ...interface{}) Template {
	var result Template
	firstArg := args[0]
	switch firstArg.(type) {
	case TableSchema:
		var items []TableSchema
		for _, table := range args {
			items = append(items, table.(TableSchema))
		}
		for _, item := range items {
			result = result.Join(CreateTable(item.Name, item.Columns).Dialect(dialect).Build())
		}
	case map[string]interface{}:
		for table, aStruct := range firstArg.(map[string]interface{}) {
			result = result.Join(CreateTable(table, aStruct).Dialect(dialect).Build())
		}
	}
	return result
}

func BatchCreateTablesIfNotExists(dialect string, args ...interface{}) Template {
	var result Template
	firstArg := args[0]
	switch firstArg.(type) {
	case TableSchema:
		var items []TableSchema
		for _, table := range args {
			items = append(items, table.(TableSchema))
		}
		for _, item := range items {
			result = result.Join(CreateTable(item.Name, item.Columns).Dialect(dialect).OnlyCreateIfNotExists(true).Build())
		}
	case map[string]interface{}:
		for table, aStruct := range firstArg.(map[string]interface{}) {
			result = result.Join(CreateTable(table, aStruct).OnlyCreateIfNotExists(true).Dialect(dialect).Build())
		}
	}
	return result
}

func BatchDropTables(tables []string, onExists bool, dialect ...string) Template {
	finalDialect := ""
	if len(dialect) > 0 {
		finalDialect = dialect[0]
	}
	var result Template
	for _, table := range tables {
		result = result.Join(NewTableBuilder().DropTable(table).OnlyDropIfExists(onExists).Dialect(finalDialect).Build())
	}
	return result
}

func BatchDropTablesIfExists(dialect string, tables ...string) Template {
	var result Template
	for _, table := range tables {
		result = result.Join(NewTableBuilder().DropTable(table).Dialect(dialect).OnlyDropIfExists(true).Build())
	}
	return result
}

func (b *TableBuilder) CreateTable(table string, columns ...interface{}) *TableBuilder {
	return b.Action(TableActionCreateTable).Table(table).Columns(columns...)
}

func (b *TableBuilder) CreateTableIfNotExists(table string, columns ...interface{}) *TableBuilder {
	return b.Action(TableActionCreateTable).Table(table).Columns(columns...).OnlyCreateIfNotExists(true)
}

func (b *TableBuilder) DropTable(table string) *TableBuilder {
	return b.Action(TableActionDropTable).Table(table)
}

func (b *TableBuilder) DropTableIfExists(table string) *TableBuilder {
	return b.Action(TableActionDropTable).Table(table).OnlyDropIfExists(true)
}

func (b *TableBuilder) Dialect(name string) *TableBuilder {
	b.dialect = name
	return b
}

func (b *TableBuilder) Action(a TableAction) *TableBuilder {
	b.action = a
	return b
}

func (b *TableBuilder) Table(name string) *TableBuilder {
	b.table = name
	return b
}

func (b *TableBuilder) Columns(columns ...interface{}) *TableBuilder {
	if len(columns) == 0 {
		return b
	}
	firstColumn := columns[0]
	switch v := firstColumn.(type) {
	case []ColumnSchema:
		b.columns = append(b.columns, v...)
	default:
		firstColumnValue := reflect.ValueOf(firstColumn)
		for firstColumnValue.Kind() == reflect.Ptr {
			firstColumnValue = firstColumnValue.Elem()
		}
		switch firstColumnValue.Kind() {
		case reflect.Struct:
			b.columns = append(b.columns, parseFields(reflect.TypeOf(firstColumn)).columns(b.dialect)...)
		default:
			panic("bear: table builder columns: unsupported args type")
		}
	}
	return b
}

func (b *TableBuilder) OnlyDropIfExists(value ...bool) *TableBuilder {
	finalValue := true
	if len(value) > 0 {
		finalValue = value[0]
	}
	b.onlyDropIfExists = finalValue
	return b
}

func (b *TableBuilder) OnlyCreateIfNotExists(value ...bool) *TableBuilder {
	finalValue := true
	if len(value) > 0 {
		finalValue = value[0]
	}
	b.onCreateIfExists = finalValue
	return b
}

func (b *TableBuilder) Indent(prefix string, indent string) *TableBuilder {
	b.newline = true
	b.prefix = prefix
	b.indent = indent
	return b
}

func (b *TableBuilder) Include(names ...string) *TableBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *TableBuilder) Exclude(names ...string) *TableBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
	return b
}

func (b *TableBuilder) Prepend(template string, values ...interface{}) *TableBuilder {
	b.prepends = append(b.prepends, NewTemplate(template, values...))
	return b
}

func (b *TableBuilder) Append(template string, values ...interface{}) *TableBuilder {
	b.appends = append(b.appends, NewTemplate(template, values...))
	return b
}

func (b *TableBuilder) PrependInner(template string, values ...interface{}) *TableBuilder {
	b.innerPrepends = append(b.innerPrepends, NewTemplate(template, values...))
	return b
}

func (b *TableBuilder) AppendInner(template string, values ...interface{}) *TableBuilder {
	b.innerAppends = append(b.innerAppends, NewTemplate(template, values...))
	return b
}

func (b *TableBuilder) Build() Template {
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
	case TableActionCreateTable:
		var columns = b.finalColumns()
		if len(columns) == 0 {
			break
		}
		if b.newline {
			if buffer.Len() > 0 && !strings.HasSuffix(buffer.String(), "\n") {
				buffer.WriteString("\n")
			}
		}
		if b.onCreateIfExists {
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
	case TableActionDropTable:
		if b.onlyDropIfExists {
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

func (b *TableBuilder) Exec(ctx context.Context, db Executor) (*Result, error) {
	return b.Build().Exec(ctx, db)
}

func (b *TableBuilder) finalColumns() []ColumnSchema {
	var includedColumns []ColumnSchema
	if len(b.include) > 0 {
		for _, column := range b.columns {
			if b.include[column.Name] {
				includedColumns = append(includedColumns, column)
			}
		}
	} else {
		includedColumns = b.columns
	}
	var excludedColumns []ColumnSchema
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

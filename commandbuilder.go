package bear

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type CommandAction string

const (
	CommandActionInsert CommandAction = "insert"
	CommandActionUpdate CommandAction = "update"
	CommandActionDelete CommandAction = "delete"
)

func (a CommandAction) Valid() bool {
	switch a {
	case CommandActionInsert, CommandActionUpdate, CommandActionDelete:
		return true
	default:
		return false
	}
}

type commandBuilder struct {
	action  CommandAction
	dialect string
	table   string
	columns []Template
	include map[string]bool
	exclude map[string]bool
	where   Condition
}

func NewCommandBuilder(dialect ...string) *commandBuilder {
	b := &commandBuilder{}
	if len(dialect) > 0 {
		b.Dialect(dialect[0])
	}
	return b
}

func Insert(table string, pairs ...map[string]interface{}) *commandBuilder {
	return NewCommandBuilder().Insert(table, pairs...)
}

func InsertStruct(table string, aStruct interface{}, includeZeroValue ...bool) *commandBuilder {
	return NewCommandBuilder().InsertStruct(table, aStruct, includeZeroValue...)
}

func Update(table string, pairs ...map[string]interface{}) *commandBuilder {
	return NewCommandBuilder().Update(table, pairs...)
}

func UpdateStruct(table string, aStruct interface{}, includeZeroValue ...bool) *commandBuilder {
	return NewCommandBuilder().UpdateStruct(table, aStruct, includeZeroValue...)
}

func Delete(table string) *commandBuilder {
	return NewCommandBuilder().Delete(table)
}

func (b *commandBuilder) Action(a CommandAction) *commandBuilder {
	if !a.Valid() {
		panic(fmt.Sprintf("bear: invalid command action %q", string(a)))
	}
	b.action = a
	return b
}

func (b *commandBuilder) Insert(table string, pairs ...map[string]interface{}) *commandBuilder {
	b.Table(table).Action(CommandActionInsert)
	if len(pairs) > 0 {
		b.SetMap(pairs[0])
	}
	return b
}

func (b *commandBuilder) InsertStruct(table string, aStruct interface{}, includeZeroValue ...bool) *commandBuilder {
	return b.Table(table).Action(CommandActionInsert).SetStruct(aStruct, includeZeroValue...)
}

func (b *commandBuilder) Update(table string, pairs ...map[string]interface{}) *commandBuilder {
	b.Table(table).Action(CommandActionUpdate)
	if len(pairs) > 0 {
		b.SetMap(pairs[0])
	}
	return b
}

func (b *commandBuilder) UpdateStruct(table string, aStruct interface{}, includeZeroValue ...bool) *commandBuilder {
	return b.Table(table).Action(CommandActionUpdate).SetStruct(aStruct, includeZeroValue...)
}

func (b *commandBuilder) Delete(table string) *commandBuilder {
	return b.Table(table).Action(CommandActionDelete)
}

func (b *commandBuilder) Dialect(name string) *commandBuilder {
	b.dialect = name
	return b
}

func (b *commandBuilder) Table(name string) *commandBuilder {
	b.table = name
	return b
}

func (b *commandBuilder) Set(column string, value interface{}) *commandBuilder {
	b.columns = append(b.columns, NewTemplate(column, value))
	return b
}

func (b *commandBuilder) SetMap(m map[string]interface{}) *commandBuilder {
	for k, v := range m {
		b.Set(k, v)
	}
	return b
}

func (b *commandBuilder) SetStruct(aStruct interface{}, includeZeroValue ...bool) *commandBuilder {
	if aStruct == nil {
		return b
	}
	finalIncludeZeroValue := false
	if len(includeZeroValue) > 0 {
		finalIncludeZeroValue = includeZeroValue[0]
	}
	m := toColumnValueMapFromStruct(reflect.ValueOf(aStruct), finalIncludeZeroValue)
	return b.SetMap(m)
}

func (b *commandBuilder) Include(names ...string) *commandBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *commandBuilder) Exclude(names ...string) *commandBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
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

func (b *commandBuilder) WhereStruct(aStruct interface{}, includeZeroValue ...bool) *commandBuilder {
	b.where = b.where.AppendStruct(aStruct, includeZeroValue...)
	return b
}

func (b *commandBuilder) Build() Template {
	result := Template{}
	columns := b.finalColumns()
	switch b.action {
	case CommandActionInsert:
		var (
			names   []string
			values  []interface{}
			holders []string
		)
		for _, column := range columns {
			names = append(names, column.Format)
			values = append(values, column.Values...)
			holders = append(holders, "?")
		}
		result.Format = fmt.Sprintf("insert into %s(%s) values(%s)",
			b.table,
			strings.Join(names, ","),
			strings.Join(holders, ","),
		)
		result.Values = append(result.Values, values...)
	case CommandActionUpdate:
		var (
			values []interface{}
			pairs  []string
		)
		for _, column := range columns {
			values = append(values, column.Values...)
			pairs = append(pairs, fmt.Sprintf("%s=?", column.Format))
		}
		result.Format = fmt.Sprintf("update %s set %s",
			b.table,
			strings.Join(pairs, ","),
		)
		result.Values = append(result.Values, values...)
	case CommandActionDelete:
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

func (b *commandBuilder) finalColumns() []Template {
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

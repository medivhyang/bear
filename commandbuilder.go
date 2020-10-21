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

type CommandBuilder struct {
	action  CommandAction
	dialect string
	table   string
	columns []Template
	include map[string]bool
	exclude map[string]bool
	where   Condition
}

func NewCommandBuilder(dialect ...string) *CommandBuilder {
	b := &CommandBuilder{}
	if len(dialect) > 0 {
		b.Dialect(dialect[0])
	}
	return b
}

func Insert(table string, pairs ...map[string]interface{}) *CommandBuilder {
	return NewCommandBuilder().Insert(table, pairs...)
}

func InsertStruct(table string, aStruct interface{}, includeZeroValue ...bool) *CommandBuilder {
	return NewCommandBuilder().InsertStruct(table, aStruct, includeZeroValue...)
}

func Update(table string, pairs ...map[string]interface{}) *CommandBuilder {
	return NewCommandBuilder().Update(table, pairs...)
}

func UpdateStruct(table string, aStruct interface{}, includeZeroValue ...bool) *CommandBuilder {
	return NewCommandBuilder().UpdateStruct(table, aStruct, includeZeroValue...)
}

func Delete(table string) *CommandBuilder {
	return NewCommandBuilder().Delete(table)
}

func (b *CommandBuilder) Action(a CommandAction) *CommandBuilder {
	if !a.Valid() {
		panic(fmt.Sprintf("bear: invalid command action %q", string(a)))
	}
	b.action = a
	return b
}

func (b *CommandBuilder) Insert(table string, pairs ...map[string]interface{}) *CommandBuilder {
	b.Table(table).Action(CommandActionInsert)
	if len(pairs) > 0 {
		b.SetMap(pairs[0])
	}
	return b
}

func (b *CommandBuilder) InsertStruct(table string, aStruct interface{}, includeZeroValue ...bool) *CommandBuilder {
	return b.Table(table).Action(CommandActionInsert).SetStruct(aStruct, includeZeroValue...)
}

func (b *CommandBuilder) Update(table string, pairs ...map[string]interface{}) *CommandBuilder {
	b.Table(table).Action(CommandActionUpdate)
	if len(pairs) > 0 {
		b.SetMap(pairs[0])
	}
	return b
}

func (b *CommandBuilder) UpdateStruct(table string, aStruct interface{}, includeZeroValue ...bool) *CommandBuilder {
	return b.Table(table).Action(CommandActionUpdate).SetStruct(aStruct, includeZeroValue...)
}

func (b *CommandBuilder) Delete(table string) *CommandBuilder {
	return b.Table(table).Action(CommandActionDelete)
}

func (b *CommandBuilder) Dialect(name string) *CommandBuilder {
	b.dialect = name
	return b
}

func (b *CommandBuilder) Table(name string) *CommandBuilder {
	b.table = name
	return b
}

func (b *CommandBuilder) Set(column string, value interface{}) *CommandBuilder {
	b.columns = append(b.columns, NewTemplate(column, value))
	return b
}

func (b *CommandBuilder) SetMap(m map[string]interface{}) *CommandBuilder {
	for k, v := range m {
		b.Set(k, v)
	}
	return b
}

func (b *CommandBuilder) SetStruct(aStruct interface{}, includeZeroValue ...bool) *CommandBuilder {
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

func (b *CommandBuilder) Include(names ...string) *CommandBuilder {
	if b.include == nil {
		b.include = map[string]bool{}
	}
	for _, name := range names {
		b.include[name] = true
	}
	return b
}

func (b *CommandBuilder) Exclude(names ...string) *CommandBuilder {
	if b.exclude == nil {
		b.exclude = map[string]bool{}
	}
	for _, name := range names {
		b.exclude[name] = true
	}
	return b
}

func (b *CommandBuilder) Where(format string, values ...interface{}) *CommandBuilder {
	b.where = b.where.AppendTemplate(Template{Format: format, Values: values})
	return b
}

func (b *CommandBuilder) WhereTemplate(templates ...Template) *CommandBuilder {
	b.where = b.where.AppendTemplate(templates...)
	return b
}

func (b *CommandBuilder) WhereMap(m map[string]interface{}) *CommandBuilder {
	b.where = b.where.AppendMap(m)
	return b
}

func (b *CommandBuilder) WhereStruct(aStruct interface{}, includeZeroValue ...bool) *CommandBuilder {
	b.where = b.where.AppendStruct(aStruct, includeZeroValue...)
	return b
}

func (b *CommandBuilder) Build() Template {
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

func (b *CommandBuilder) Execute(db DB) (*Result, error) {
	return b.Build().Execute(db)
}

func (b *CommandBuilder) ExecuteContext(ctx context.Context, db DB) (*Result, error) {
	return b.Build().ExecuteContext(ctx, db)
}

func (b *CommandBuilder) finalColumns() []Template {
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

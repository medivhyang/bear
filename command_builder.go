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

func Insert(table string, args ...interface{}) *CommandBuilder {
	return NewCommandBuilder().Insert(table, args...)
}

func Update(table string, args ...interface{}) *CommandBuilder {
	return NewCommandBuilder().Update(table, args...)
}

func Delete(table string) *CommandBuilder {
	return NewCommandBuilder().Delete(table)
}

func (b *CommandBuilder) Action(a CommandAction) *CommandBuilder {
	if !a.Valid() {
		panic(fmt.Sprintf("bear: command builder action: invalid action %q", string(a)))
	}
	b.action = a
	return b
}

func (b *CommandBuilder) Insert(table string, columns ...interface{}) *CommandBuilder {
	b.Table(table).Action(CommandActionInsert)
	if len(columns) == 0 {
		return b
	}
	firstColumn := columns[0]
	firstColumnValue := reflect.ValueOf(firstColumn)
	switch firstColumnValue.Kind() {
	case reflect.Map:
		b.setMap(firstColumn.(map[string]interface{}))
	case reflect.Struct:
		if len(columns) == 2 {
			b.setStruct(firstColumn, columns[1].(bool))
		} else {
			b.setStruct(firstColumn, false)
		}
	default:
		panic("bear: command builder insert: unsupported args type")
	}
	return b
}

func (b *CommandBuilder) Update(table string, columns ...interface{}) *CommandBuilder {
	b.Table(table).Action(CommandActionUpdate)
	if len(columns) == 0 {
		return b
	}
	firstColumn := columns[0]
	firstColumnValue := reflect.ValueOf(firstColumn)
	switch firstColumnValue.Kind() {
	case reflect.Map:
		b.setMap(firstColumn.(map[string]interface{}))
	case reflect.Struct:
		if len(columns) == 2 {
			b.setStruct(firstColumn, columns[1].(bool))
		} else {
			b.setStruct(firstColumn, false)
		}
	default:
		panic("bear: command builder update: unsupported args type")
	}
	return b
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

func (b *CommandBuilder) Set(args ...interface{}) *CommandBuilder {
	if len(args) == 0 {
		return b
	}
	firstArg := args[0]
	firstArgValue := reflect.ValueOf(firstArg)
	switch firstArgValue.Kind() {
	case reflect.String:
		if len(args) != 2 {
			panic("bear: command builder set: unsupported args type")
		}
		b.set(firstArg.(string), args[1])
	case reflect.Map:
		b.setMap(firstArg.(map[string]interface{}))
	case reflect.Struct:
		if len(args) >= 2 {
			b.setStruct(firstArg, args[1].(bool))
		} else {
			b.setStruct(firstArg, false)
		}
	default:
		panic("bear: command builder set: unsupported args type")
	}
	return b
}

func (b *CommandBuilder) set(column string, value interface{}) *CommandBuilder {
	b.columns = append(b.columns, NewTemplate(column, value))
	return b
}

func (b *CommandBuilder) setMap(m map[string]interface{}) *CommandBuilder {
	for k, v := range m {
		b.Set(k, v)
	}
	return b
}

func (b *CommandBuilder) setStruct(aStruct interface{}, includeZeroValue bool) *CommandBuilder {
	if aStruct == nil {
		return b
	}
	m := parseColumnValueMap(reflect.ValueOf(aStruct), includeZeroValue)
	return b.setMap(m)
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

func (b *CommandBuilder) Where(args ...interface{}) *CommandBuilder {
	b.where = b.where.Append(args...)
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

func (b *CommandBuilder) Exec(ctx context.Context, db DB) (*Result, error) {
	return b.Build().Exec(ctx, db)
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

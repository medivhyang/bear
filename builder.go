package bear

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const (
	ActionSelect = "select"
	ActionInsert = "insert"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

type Segment struct {
	Template string
	Values   []interface{}
}

func New(template string, values ...interface{}) *Segment {
	return &Segment{Template: template, Values: values}
}

func (s Segment) Tuple() (string, []interface{}) {
	return s.Template, s.Values
}

func Select(table string, fields ...string) *QueryBuilder {
	b := QueryBuilder{table: table}
	for _, field := range fields {
		b.fields = append(b.fields, Segment{Template: field})
	}
	return &b
}

func Insert(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: ActionInsert}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func Update(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: ActionUpdate}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func Delete(table string) *CommandBuilder {
	b := &CommandBuilder{table: table, action: ActionDelete}
	return b
}

type QueryBuilder struct {
	table   string
	fields  []Segment
	where   ConditionBuilder
	groupBy []string
	having  ConditionBuilder
	orderBy []string
}

func (b *QueryBuilder) Fields(fields ...Segment) *QueryBuilder {
	b.fields = append(b.fields, fields...)
	return b
}

func (b *QueryBuilder) Where(template string, values ...interface{}) *QueryBuilder {
	b.where.Append(Segment{Template: template, Values: values})
	return b
}

func (b *QueryBuilder) WhereWithTuple(template string, values []interface{}) *QueryBuilder {
	b.where.Append(Segment{Template: template, Values: values})
	return b
}

func (b *QueryBuilder) Having(template string, values ...interface{}) *QueryBuilder {
	b.having.Append(Segment{Template: template, Values: values})
	return b
}

func (b *QueryBuilder) HavingWithTuple(template string, values []interface{}) *QueryBuilder {
	b.having.Append(Segment{Template: template, Values: values})
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

func (b *QueryBuilder) mappingFields() []string {
	var result []string
	for _, field := range b.fields {
		splits := strings.Split(field.Template, " ")
		result = append(result, splits[len(splits)-1])
	}
	return result
}

func (b *QueryBuilder) Build() Segment {
	result := Segment{}

	var fieldsTemplates []string
	for _, field := range b.fields {
		fieldsTemplates = append(fieldsTemplates, field.Template)
		result.Values = append(result.Values, field.Values...)
	}
	result.Template = fmt.Sprintf("select %s from %s",
		strings.Join(fieldsTemplates, ","),
		b.table,
	)

	where := b.where.Build()
	if len(where.Template) > 0 {
		result.Template += " where " + where.Template
		result.Values = append(result.Values, where.Values...)
	}

	if len(b.groupBy) > 0 {
		result.Template += fmt.Sprintf(" group by %s", strings.Join(b.groupBy, ","))
	}

	having := b.having.Build()
	if len(having.Template) > 0 {
		result.Template += " having" + having.Template
		result.Values = append(result.Values, having.Values...)
	}

	if len(b.orderBy) > 0 {
		result.Template += fmt.Sprintf(" order by %s", strings.Join(b.orderBy, ","))
	}

	return result
}

func (b *QueryBuilder) As(alias string) Segment {
	s := b.Build()
	s.Template = fmt.Sprintf("(%s) as %s", s.Template, alias)
	return s
}

type Querier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func (b *QueryBuilder) Query(querier Querier) *QueryResult {
	s := b.Build()
	return WrapQueryResult(querier.Query(s.Template, s.Values...))
}

type WithContextQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func (b *QueryBuilder) QueryWithContext(querier WithContextQuerier, ctx context.Context) *QueryResult {
	s := b.Build()
	return WrapQueryResult(querier.QueryContext(ctx, s.Template, s.Values...))
}

type CommandBuilder struct {
	action string
	table  string
	names  []string
	values []interface{}
	where  ConditionBuilder
}

func (b *CommandBuilder) Where(template string, values ...interface{}) *CommandBuilder {
	b.where.Append(Segment{Template: template, Values: values})
	return b
}

func (b *CommandBuilder) WhereWithTuple(template string, values []interface{}) *CommandBuilder {
	b.where.Append(Segment{Template: template, Values: values})
	return b
}

func (b *CommandBuilder) Build() Segment {
	result := Segment{}

	switch b.action {
	case ActionInsert:
		var holders []string
		for i := 0; i < len(b.values); i++ {
			holders = append(holders, "?")
		}
		result.Template = fmt.Sprintf("insert into %s(%s) values(%s)",
			b.table,
			strings.Join(b.names, ","),
			strings.Join(holders, ","),
		)
		result.Values = append(result.Values, b.values...)
	case ActionUpdate:
		var pairs []string
		for i := 0; i < len(b.names); i++ {
			pairs = append(pairs, fmt.Sprintf("%s=?", b.names[i]))
		}
		result.Template = fmt.Sprintf("update %s set %s",
			b.table,
			strings.Join(pairs, ","),
		)
		result.Values = append(result.Values, b.values...)
	case ActionDelete:
		result.Template = fmt.Sprintf("delete from %s", b.table)
	}

	where := b.where.Build()
	if len(where.Template) > 0 {
		result.Template += " where" + where.Template
		result.Values = append(result.Values, where.Values...)
	}

	return result
}

type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (b *CommandBuilder) Execute(exectutor Executor) *CommandResult {
	s := b.Build()
	return WrapCommandResult(exectutor.Exec(s.Template, s.Values...))
}

type WithContextExectutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (b *CommandBuilder) ExecuteWithContext(exectutor WithContextExectutor, ctx context.Context) *CommandResult {
	s := b.Build()
	return WrapCommandResult(exectutor.ExecContext(ctx, s.Template, s.Values...))
}

type ConditionBuilder struct {
	conditions []string
	names      []string
	values     []interface{}
}

func (b *ConditionBuilder) Append(items ...Segment) *ConditionBuilder {
	for _, item := range items {
		b.conditions = append(b.conditions, item.Template)
		b.values = append(b.values, item.Values...)
	}
	return b
}

func (b *ConditionBuilder) Build() Segment {
	if len(b.conditions) == 0 {
		return Segment{}
	}

	names := make([]string, len(b.names))
	copy(names, b.names)

	values := make([]interface{}, len(b.values))
	copy(values, b.values)

	var conditions []string
	for _, condition := range b.conditions {
		conditions = append(conditions, "("+condition+")")
	}

	result := Segment{
		Template: strings.Join(conditions, " and "),
		Values:   values,
	}

	return result
}

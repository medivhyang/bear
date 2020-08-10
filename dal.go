package bear

import (
	"context"
	"fmt"
	"strings"
)

const (
	actionSelect = "select"
	actionInsert = "insert"
	actionUpdate = "update"
	actionDelete = "delete"
)

func Select(table string, fields ...string) *QueryBuilder {
	b := QueryBuilder{table: table}
	for _, field := range fields {
		b.fields = append(b.fields, Segment{Template: field})
	}
	return &b
}

func Insert(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: actionInsert}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func Update(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: actionUpdate}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func Delete(table string) *CommandBuilder {
	b := &CommandBuilder{table: table, action: actionDelete}
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

func (b *QueryBuilder) Where(template string, values []interface{}) *QueryBuilder {
	b.where.Append(Segment{Template: template, Values: values})
	return b
}

func (b *QueryBuilder) Having(template string, values []interface{}) *QueryBuilder {
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

func (b *QueryBuilder) Query(querier Querier) (*Rows, error) {
	s := b.Build()
	rows, err := querier.Query(s.Template, s.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

func (b *QueryBuilder) QueryWithContext(querier WithContextQuerier, ctx context.Context) (*Rows, error) {
	s := b.Build()
	rows, err := querier.QueryContext(ctx, s.Template, s.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

type CommandBuilder struct {
	action string
	table  string
	names  []string
	values []interface{}
	where  ConditionBuilder
}

func (b *CommandBuilder) Where(template string, values []interface{}) *CommandBuilder {
	b.where.Append(Segment{Template: template, Values: values})
	return b
}

func (b *CommandBuilder) Build() Segment {
	result := Segment{}

	switch b.action {
	case actionInsert:
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
	case actionUpdate:
		var pairs []string
		for i := 0; i < len(b.names); i++ {
			pairs = append(pairs, fmt.Sprintf("%s=?", b.names[i]))
		}
		result.Template = fmt.Sprintf("update %s set %s",
			b.table,
			strings.Join(pairs, ","),
		)
		result.Values = append(result.Values, b.values...)
	case actionDelete:
		result.Template = fmt.Sprintf("delete from %s", b.table)
	}

	where := b.where.Build()
	if len(where.Template) > 0 {
		result.Template += " where" + where.Template
		result.Values = append(result.Values, where.Values...)
	}

	return result
}

func (b *CommandBuilder) Execute(exectutor Executor) (*Result, error) {
	s := b.Build()
	result, err := exectutor.Exec(s.Template, s.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

func (b *CommandBuilder) ExecuteWithContext(exectutor WithContextExectutor, ctx context.Context) (*Result, error) {
	s := b.Build()
	result, err := exectutor.ExecContext(ctx, s.Template, s.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
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

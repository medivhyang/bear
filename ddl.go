package bear

import (
	"context"
)

func CreateTable(table string, fields []DBField) *TableBuilder {
	return &TableBuilder{
		table:  table,
		fields: fields,
	}
}

type DBField struct {
	Name   string
	Type   string
	PK     bool
	Auto   bool
	Suffix string
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
	table  string
	fields []DBField
}

func (b *TableBuilder) Build() *Segment {
	result := Segment{}
	return &result
}

func (b *TableBuilder) Execute(executor Executor) (*Result, error) {
	s := b.Build()
	result, err := executor.Exec(s.Template, s.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

func (b *TableBuilder) ExecuteContext(executor WithContextExectutor, ctx context.Context) (*Result, error) {
	s := b.Build()
	result, err := executor.ExecContext(ctx, s.Template, s.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

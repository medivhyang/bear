package bear

import (
	"context"
	"errors"
)

type Template struct {
	Format string
	Values []interface{}
}

func New(format string, values ...interface{}) Template {
	return Template{Format: format, Values: values}
}

func Empty() Template {
	return Template{}
}

func (t Template) Append(format string, values ...interface{}) Template {
	t.Format += format
	t.Values = append(t.Values, values...)
	return t
}

func (t Template) Merge(other Template) Template {
	return t.Append(other.Format, other.Values...)
}

func (t Template) Join(other Template, sep string) Template {
	return t.Append(sep+other.Format, other.Values...)
}

func (t Template) And(other Template) Template {
	if other.Format == "" {
		return t
	}
	if t.Format == "" {
		return other
	}
	return t.Join(other, " and ")
}

func (t Template) Or(other Template) Template {
	if other.Format == "" {
		return t
	}
	if t.Format == "" {
		return other
	}
	return t.Join(other, " or ")
}

func (t Template) Query(querier Querier) (*Rows, error) {
	if t.IsEmpty() {
		return nil, errors.New("bear: empty template")
	}
	rows, err := querier.Query(t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

func (t Template) QueryWithContext(querier WithContextQuerier, ctx context.Context) (*Rows, error) {
	if t.IsEmpty() {
		return nil, errors.New("bear: empty template")
	}
	rows, err := querier.QueryContext(ctx, t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

func (t Template) Execute(executor Executor) (*Result, error) {
	if t.IsEmpty() {
		return nil, errors.New("bear: empty template")
	}
	result, err := executor.Exec(t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

func (t Template) ExecuteWitchContext(executor WithContextExectutor, ctx context.Context) (*Result, error) {
	if t.IsEmpty() {
		return nil, errors.New("bear: empty template")
	}
	result, err := executor.ExecContext(ctx, t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

func (t Template) IsEmpty() bool {
	return t.Format == "" && len(t.Values) == 0
}

package bear

import "context"

type Template struct {
	Format string
	Values []interface{}
}

func New(format string, values ...interface{}) Template {
	return Template{Format: format, Values: values}
}

func (t Template) Append(format string, values ...interface{}) Template {
	t.Format += format
	t.Values = append(t.Values, values...)
	return t
}

func (t Template) Merge(other Template) Template {
	return t.Append(other.Format, other.Values...)
}

func (t Template) Query(querier Querier) (*Rows, error) {
	rows, err := querier.Query(t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

func (t Template) QueryWithContext(querier WithContextQuerier, ctx context.Context) (*Rows, error) {
	rows, err := querier.QueryContext(ctx, t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

func (t Template) Execute(executor Executor) (*Result, error) {
	result, err := executor.Exec(t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

func (t Template) ExecuteWitchContext(executor WithContextExectutor, ctx context.Context) (*Result, error) {
	result, err := executor.ExecContext(ctx, t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

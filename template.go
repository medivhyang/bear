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

func (t Template) Prepend(format string, values ...interface{}) Template {
	t.Format = format + t.Format
	if len(values) > 0 {
		t.Values = append(append([]interface{}{}, values...), t.Values...)
	}
	return t
}

func (t Template) Append(format string, values ...interface{}) Template {
	t.Format += format
	if len(values) > 0 {
		t.Values = append(t.Values, values...)
	}
	return t
}

func (t Template) Wrap(left string, right string) Template {
	return t.Prepend(left).Append(right)
}

func (t Template) WrapBracket() Template {
	return t.Wrap("(", ")")
}

func (t Template) Join(other Template, sep ...string) Template {
	finalSep := ""
	if len(sep) > 0 {
		finalSep = sep[0]
	}
	return t.Append(finalSep+other.Format, other.Values...)
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

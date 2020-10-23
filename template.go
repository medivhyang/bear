package bear

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrEmptyTemplate = errors.New("bear: empty template")

type Template struct {
	Format string
	Values []interface{}
}

func NewTemplate(format string, values ...interface{}) Template {
	return Template{Format: format, Values: values}
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
	var newFormat string
	if t.IsEmptyOrWhitespace() {
		newFormat = other.Format
	} else if other.IsEmptyOrWhitespace() {
		newFormat = t.Format
	} else {
		newFormat = t.Format + finalSep + other.Format
	}
	newValues := append(t.Values, other.Values...)
	return NewTemplate(newFormat, newValues...)
}

func (t Template) And(other Template) Template {
	return t.Join(other, " and ")
}

func (t Template) Or(other Template) Template {
	return t.Join(other, " or ")
}

func (t Template) Query(db DB) (*Rows, error) {
	if t.IsEmptyOrWhitespace() {
		return nil, ErrEmptyTemplate
	}
	debugf("query: %s\n", t)
	rows, err := db.Query(t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

func (t Template) QueryContext(ctx context.Context, db DB) (*Rows, error) {
	if t.IsEmptyOrWhitespace() {
		return nil, ErrEmptyTemplate
	}
	debugf("query: %s\n", t)
	rows, err := db.QueryContext(ctx, t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapRows(rows), nil
}

func (t Template) QueryScalar(ctx context.Context, db DB, value interface{}) error {
	rows, err := t.QueryContext(ctx, db)
	if err != nil {
		return err
	}
	return rows.Scalar(value)
}

func (t Template) QueryScalarSlice(ctx context.Context, db DB, values interface{}) error {
	rows, err := t.QueryContext(ctx, db)
	if err != nil {
		return err
	}
	return rows.ScalarSlice(values)
}

func (t Template) QueryMap(ctx context.Context, db DB) (map[string]interface{}, error) {
	rows, err := t.QueryContext(ctx, db)
	if err != nil {
		return nil, err
	}
	return rows.Map()
}

func (t Template) QueryMapSlice(ctx context.Context, db DB) ([]map[string]interface{}, error) {
	rows, err := t.QueryContext(ctx, db)
	if err != nil {
		return nil, err
	}
	return rows.MapSlice()
}

func (t Template) QueryStruct(ctx context.Context, db DB, structPtr interface{}) error {
	rows, err := t.QueryContext(ctx, db)
	if err != nil {
		return err
	}
	return rows.Struct(structPtr)
}

func (t Template) QueryStructSlice(ctx context.Context, db DB, structPtr interface{}) error {
	rows, err := t.QueryContext(ctx, db)
	if err != nil {
		return err
	}
	return rows.StructSlice(structPtr)
}

func (t Template) Execute(db DB) (*Result, error) {
	if t.IsEmptyOrWhitespace() {
		return nil, ErrEmptyTemplate
	}
	debugf("exec: %s\n", t)
	result, err := db.Exec(t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

func (t Template) ExecuteContext(ctx context.Context, db DB) (*Result, error) {
	if t.IsEmptyOrWhitespace() {
		return nil, ErrEmptyTemplate
	}
	debugf("exec: %s\n", t)
	result, err := db.ExecContext(ctx, t.Format, t.Values...)
	if err != nil {
		return nil, err
	}
	return WrapResult(result), nil
}

func (t Template) IsEmpty() bool {
	return t.Format == ""
}

func (t Template) IsEmptyOrWhitespace() bool {
	return strings.TrimSpace(t.Format) == ""
}

func (t Template) String() string {
	buffer := strings.Builder{}
	buffer.WriteString(fmt.Sprintf("%q", t.Format))
	if len(t.Values) == 0 {
		return buffer.String()
	}
	buffer.WriteString(": ")
	buffer.WriteString("[")
	for index, value := range t.Values {
		if index > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("%#v", value))
	}
	buffer.WriteString("]")
	return buffer.String()
}

package bear

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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

func (t Template) Query(ctx context.Context, db Executor, value interface{}) error {
	switch v := value.(type) {
	case *map[string]interface{}:
		return t.queryMap(ctx, db, v)
	default:
		reflectValue := reflect.ValueOf(value)
		for reflectValue.Kind() == reflect.Ptr {
			reflectValue = reflectValue.Elem()
		}
		switch reflectValue.Kind() {
		case reflect.Struct:
			return t.queryStruct(ctx, db, value)
		case reflect.Slice:
			switch reflectValue.Type().Elem().Kind() {
			case reflect.Map:
				return t.queryMapSlice(ctx, db, value)
			case reflect.Struct:
				return t.queryStructSlice(ctx, db, value)
			default:
				return t.queryScalarSlice(ctx, db, value)
			}
		default:
			return t.queryScalar(ctx, db, value)
		}
	}
}

func (t Template) QueryRows(ctx context.Context, db Executor) (*Rows, error) {
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

func (t Template) queryScalar(ctx context.Context, db Executor, value interface{}) error {
	rows, err := t.QueryRows(ctx, db)
	if err != nil {
		return err
	}
	return rows.Scalar(value)
}

func (t Template) queryScalarSlice(ctx context.Context, db Executor, values interface{}) error {
	rows, err := t.QueryRows(ctx, db)
	if err != nil {
		return err
	}
	return rows.ScalarSlice(values)
}

func (t Template) queryMap(ctx context.Context, db Executor, value *map[string]interface{}) error {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	if !reflectValue.CanSet() {
		return errors.New("bear: template query map: can't set value")
	}
	rows, err := t.QueryRows(ctx, db)
	if err != nil {
		return err
	}
	m, err := rows.Map()
	if err != nil {
		return err
	}
	reflectValue.Set(reflect.ValueOf(m))
	return nil
}

func (t Template) queryMapSlice(ctx context.Context, db Executor, value interface{}) error {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	if !reflectValue.CanSet() {
		return errors.New("bear: template query map slice: can't set value")
	}
	rows, err := t.QueryRows(ctx, db)
	if err != nil {
		return err
	}
	slice, err := rows.MapSlice()
	if err != nil {
		return err
	}
	reflectValue.Set(reflect.ValueOf(slice))
	return nil
}

func (t Template) queryStruct(ctx context.Context, db Executor, value interface{}) error {
	rows, err := t.QueryRows(ctx, db)
	if err != nil {
		return err
	}
	return rows.Struct(value)
}

func (t Template) queryStructSlice(ctx context.Context, db Executor, value interface{}) error {
	rows, err := t.QueryRows(ctx, db)
	if err != nil {
		return err
	}
	return rows.StructSlice(value)
}

func (t Template) Exec(ctx context.Context, db Executor) (*Result, error) {
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
	if len(t.Values) > 0 {
		buffer.WriteString(": ")
		buffer.WriteString(fmt.Sprintf("%#v", t.Values))
	}
	return buffer.String()
}

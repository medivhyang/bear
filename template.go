package bear

import (
	"fmt"
	"strings"
)

type Template struct {
	Format string
	Values []interface{}
}

func T(format string, values ...interface{}) Template {
	return NewTemplate(format, values...)
}

func NewTemplate(format string, values ...interface{}) Template {
	return Template{Format: format, Values: append([]interface{}{}, values...)}
}

func (t Template) Append(others ...Template) Template {
	t2 := NewTemplate(t.Format, t.Values...)
	for _, o := range others {
		t2.Format += o.Format
		t2.AppendValues(o.Values...)
	}
	return t
}

func (t Template) Appendf(format string, values ...interface{}) Template {
	return NewTemplate(t.Format, t.AppendValues(values...))
}

func (t Template) AppendValues(values ...interface{}) Template {
	t.Values = append(append([]interface{}{}, t.Values...), values...)
	return t
}

func (t Template) Wrap(left string, right string) Template {
	t.Format = left + t.Format + right
	return t
}

func (t Template) Bracket() Template {
	return t.Wrap("(", ")")
}

func (t Template) String() string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%q", t.Format))
	if len(t.Values) > 0 {
		b.WriteString(": ")
		b.WriteString(fmt.Sprintf("%#v", t.Values))
	}
	return b.String()
}

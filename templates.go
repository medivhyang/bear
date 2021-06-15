package bear

import (
	"fmt"
	"strings"
)

type Template struct {
	Format string
	Values []interface{}
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
	return NewTemplate(format, t.AppendValues(values...))
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

type Templates []Template

func NewTemplates(tt ...Template) Templates {
	return tt
}

func PlainTemplates(ss ...string) Templates {
	tt := make([]Template, 0, len(ss))
	for _, s := range ss {
		tt = append(tt, NewTemplate(s))
	}
	return tt
}

func (tt Templates) Append(others ...Template) Templates {
	return append(tt, others...)
}

func (tt Templates) Appendf(format string, values ...interface{}) Templates {
	return tt.Append(NewTemplate(format, values...))
}

func (tt Templates) Join(sep string, right, left string) Template {
	if len(tt) == 0 {
		return Template{}
	}
	ff := make([]string, 0, len(tt))
	vv := make([]interface{}, 0, len(tt))
	for _, c := range tt {
		ff = append(ff, c.Format)
		vv = append(vv, c.Values...)
	}
	return NewTemplate(right+strings.Join(ff, sep)+left, vv...)
}

func (tt Templates) Formats() []string {
	r := make([]string, 0, len(tt))
	for _, t := range tt {
		r = append(r, t.Format)
	}
	return r
}

func (tt Templates) Values() []interface{} {
	var r []interface{}
	for _, t := range tt {
		r = append(r, t.Values...)
	}
	return r
}

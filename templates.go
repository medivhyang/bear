package bear

import (
	"strings"
)

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

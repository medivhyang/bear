package expr

import (
	"fmt"
	"github.com/medivhyang/bear"
	"strings"
)

type Expr bear.Template

func New(template string, values []interface{}) Expr {
	return Expr{
		Format: template,
		Values: values,
	}
}

func Empty() Expr {
	return Expr{}
}

func (e Expr) Template() bear.Template {
	return bear.Template(e)
}

func Value(values ...interface{}) Expr {
	return Expr{
		Values: values,
	}
}

func (e Expr) And(other Expr) Expr {
	if len(other.Format) == 0 {
		return e
	}
	if len(e.Format) == 0 {
		return other
	}
	return Expr{
		Format: e.Format + " and " + other.Format,
		Values: append(e.Values, other.Values...),
	}
}

func (e Expr) Or(other Expr) Expr {
	if len(other.Format) == 0 {
		return e
	}
	if len(e.Format) == 0 {
		return other
	}
	return Expr{
		Format: "(" + e.Format + " or " + other.Format + ")",
		Values: append(e.Values, other.Values...),
	}
}

func Equal(name string, value Expr) Expr {
	result := Expr{}
	result.Values = append(result.Values, value.Values...)
	if len(value.Format) == 0 {
		result.Format = fmt.Sprintf("%s = ?", name)
		return result
	}
	result.Format = fmt.Sprintf("%s = (%s)", name, value.Format)
	return result
}

func NotEqual(name string, value Expr) Expr {
	result := Expr{}
	result.Values = append(result.Values, value.Values...)
	if len(value.Format) == 0 {
		result.Format = fmt.Sprintf("%s != ?", name)
	} else {
		result.Format = fmt.Sprintf("%s != (%s)", name, value.Format)
	}
	return result
}

func In(name string, value Expr) Expr {
	result := Expr{}
	var holder []string
	for i := 0; i < len(value.Values); i++ {
		holder = append(holder, "?")
	}
	result.Values = append(result.Values, value.Values)
	if len(value.Format) == 0 {
		result.Format = fmt.Sprintf("%s in (%s)", name, strings.Join(holder, ","))
	} else {
		result.Format = fmt.Sprintf("%s in (%s)", name, value.Format)
	}
	return result
}

func Between(name string, v1 Expr, v2 Expr) Expr {
	result := Expr{}
	result.Values = append(result.Values, v1.Values...)
	result.Values = append(result.Values, v2.Values...)
	if len(v1.Format) == 0 {
		result.Format = fmt.Sprintf("%s between ?", name)
	} else {
		result.Format = fmt.Sprintf("%s between (%s)", name, v1.Format)
	}
	if len(v2.Format) == 0 {
		result.Format += fmt.Sprintf(" and ?")
	} else {
		result.Format += fmt.Sprintf(" and (%s)", v2.Format)
	}
	return result
}

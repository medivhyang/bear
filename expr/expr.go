package expr

import (
	"fmt"
	"strings"
)

type Expr struct {
	Template string
	Values   []interface{}
}

func New(template string, values ...interface{}) Expr {
	return Expr{
		Template: template,
		Values:   values,
	}
}

func Empty() Expr {
	return Expr{ }
}

func Value(values ...interface{}) Expr {
	return Expr{
		Values: values,
	}
}

func (e Expr) Tuple() (string, []interface{}) {
	return e.Template, e.Values
}

func (e Expr) And(other Expr) Expr {
	if len(other.Template) == 0 {
		return e
	}
	if len(e.Template) == 0 {
		return other
	}
	return Expr{
		Template: e.Template + " and " + other.Template,
		Values:   append(e.Values, other.Values...),
	}
}

func (e Expr) Or(other Expr) Expr {
	if len(other.Template) == 0 {
		return e
	}
	if len(e.Template) == 0 {
		return other
	}
	return Expr{
		Template: "(" + e.Template + " or " + other.Template + ")",
		Values:   append(e.Values, other.Values...),
	}
}

func Equal(name string, value Expr) Expr {
	result := Expr{}
	result.Values = append(result.Values, value.Values...)
	if len(value.Template) == 0 {
		result.Template = fmt.Sprintf("%s = ?", name)
		return result
	}
	result.Template = fmt.Sprintf("%s = (%s)", name, value.Template)
	return result
}

func NotEqual(name string, value Expr) Expr {
	result := Expr{}
	result.Values = append(result.Values, value.Values...)
	if len(value.Template) == 0 {
		result.Template = fmt.Sprintf("%s != ?", name)
	} else {
		result.Template = fmt.Sprintf("%s != (%s)", name, value.Template)
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
	if len(value.Template) == 0 {
		result.Template = fmt.Sprintf("%s in (%s)", name, strings.Join(holder, ","))
	} else {
		result.Template = fmt.Sprintf("%s in (%s)", name, value.Template)
	}
	return result
}

func Between(name string, v1 Expr, v2 Expr) Expr {
	result := Expr{}
	result.Values = append(result.Values, v1.Values...)
	result.Values = append(result.Values, v2.Values...)
	if len(v1.Template) == 0 {
		result.Template = fmt.Sprintf("%s between ?", name)
	} else {
		result.Template = fmt.Sprintf("%s between (%s)", name, v1.Template)
	}
	if len(v2.Template) == 0 {
		result.Template += fmt.Sprintf(" and ?")
	} else {
		result.Template += fmt.Sprintf(" and (%s)", v2.Template)
	}
	return result
}

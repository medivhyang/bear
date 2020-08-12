package expr

import (
	"fmt"
	"github.com/medivhyang/bear"
	"strings"
)

func New(format string, values ...interface{}) bear.Template {
	return bear.Template{Format: format, Values: values}
}

func Empty() bear.Template {
	return bear.Template{}
}

func Value(values ...interface{}) bear.Template {
	return bear.New("", values...)
}

func Equal(name string, value interface{}) bear.Template {
	return bear.New(fmt.Sprintf("%s = ?", name), value)
}

func EqualTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s = (%s)", name, template.Format), template.Values...)
}

func NotEqual(name string, value interface{}) bear.Template {
	return bear.New(fmt.Sprintf("%s != ?", name), value)
}

func NotEqualTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s != (%s)", name, template.Format), template.Values...)
}

func LessThan(name string, value interface{}) bear.Template {
	return bear.New(fmt.Sprintf("%s < ?", name), value)
}

func LessThanTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s < (%s)", name, template.Format), template.Values...)
}

func LessEqual(name string, value interface{}) bear.Template {
	return bear.New(fmt.Sprintf("%s <= ?", name), value)
}

func LessEqualTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s <= (%s)", name, template.Format), template.Values...)
}

func GreaterThan(name string, value interface{}) bear.Template {
	return bear.New(fmt.Sprintf("%s > ?", name), value)
}

func GreaterThanTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s > (%s)", name, template.Format), template.Values...)
}

func GreaterEqual(name string, value interface{}) bear.Template {
	return bear.New(fmt.Sprintf("%s >= ?", name), value)
}

func GreaterEqualTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s >= (%s)", name, template.Format), template.Values...)
}

func Between(name string, v1 interface{}, v2 interface{}) bear.Template {
	return bear.New(fmt.Sprintf("%s between ? and ?", name), v1, v2)
}

func BetweenTemplate(name string, v1 bear.Template, v2 bear.Template) bear.Template {
	result := bear.Template{}
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

func In(name string, values ...interface{}) bear.Template {
	if len(values) == 0 {
		return Empty()
	}
	var placeholders []string
	for i := 0; i < len(values); i++ {
		placeholders = append(placeholders, "?")
	}
	return bear.New(fmt.Sprintf("%s in (%s)", name, strings.Join(placeholders, ",")), values...)
}

func InTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s in (%s)", name, template.Format), template.Values...)
}

func Like(name string, value string) bear.Template {
	return bear.New(fmt.Sprintf("%s like ?", name), value)
}

func LikeTemplate(name string, template bear.Template) bear.Template {
	return bear.New(fmt.Sprintf("%s like (%s)", name, template.Format), template.Values...)
}

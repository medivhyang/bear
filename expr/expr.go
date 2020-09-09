package expr

import (
	"fmt"
	"github.com/medivhyang/bear"
	"strings"
)

func New(format string, values ...interface{}) bear.Template {
	return bear.Template{Format: format, Values: values}
}

func Value(values ...interface{}) bear.Template {
	return bear.NewTemplate("", values...)
}

func EQ(name string, value interface{}) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s = ?", name), value)
}

func EQTemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s = (%s)", name, template.Format), template.Values...)
}

func NE(name string, value interface{}) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s != ?", name), value)
}

func NETemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s != (%s)", name, template.Format), template.Values...)
}

func LT(name string, value interface{}) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s < ?", name), value)
}

func LTTemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s < (%s)", name, template.Format), template.Values...)
}

func LE(name string, value interface{}) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s <= ?", name), value)
}

func LETemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s <= (%s)", name, template.Format), template.Values...)
}

func GT(name string, value interface{}) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s > ?", name), value)
}

func GTTemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s > (%s)", name, template.Format), template.Values...)
}

func GE(name string, value interface{}) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s >= ?", name), value)
}

func GETemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s >= (%s)", name, template.Format), template.Values...)
}

func Between(name string, v1 interface{}, v2 interface{}) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s between ? and ?", name), v1, v2)
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
	if name == "" || len(values) == 0 {
		return bear.Template{}
	}
	var placeholders []string
	for i := 0; i < len(values); i++ {
		placeholders = append(placeholders, "?")
	}
	return bear.NewTemplate(fmt.Sprintf("%s in (%s)", name, strings.Join(placeholders, ",")), values...)
}

func InTemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s in (%s)", name, template.Format), template.Values...)
}

func Like(name string, value string) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s like ?", name), value)
}

func LikeTemplate(name string, template bear.Template) bear.Template {
	return bear.NewTemplate(fmt.Sprintf("%s like (%s)", name, template.Format), template.Values...)
}

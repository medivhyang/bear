package expr

import (
	"fmt"
	"github.com/medivhyang/bear"
	"reflect"
	"strings"
)

func New(format string, values ...interface{}) bear.Template {
	return bear.NewTemplate(format, values...)
}

func Value(values ...interface{}) bear.Template {
	return bear.NewTemplate("", values...)
}

func Equal(name string, value interface{}) bear.Template {
	return binary("=", name, value)
}

func NotEqual(name string, value interface{}) bear.Template {
	return binary("!=", name, value)
}

func LessThan(name string, value interface{}) bear.Template {
	return binary("<", name, value)
}

func LessEqual(name string, value interface{}) bear.Template {
	return binary("<=", name, value)
}

func GreaterThan(name string, value interface{}) bear.Template {
	return binary(">", name, value)
}

func GreaterEqual(name string, value interface{}) bear.Template {
	return binary(">=", name, value)
}

func Like(name string, value string) bear.Template {
	return binary("like", name, value)
}

func binary(op string, name string, value interface{}) bear.Template {
	var (
		format string
		values []interface{}
	)
	t, ok := value.(bear.Template)
	if ok {
		if t.IsEmptyOrWhitespace() {
			format = fmt.Sprintf("%s %s ?", name, op)
		} else {
			format = fmt.Sprintf("%s %s (%s)", name, op, t.Format)
		}
		value = append(values, t.Values...)
	} else {
		format = fmt.Sprintf("%s %s ?", name, op)
		values = append(values, value)
	}
	return bear.NewTemplate(format, values...)
}

func In(name string, value interface{}) bear.Template {
	return in(name, value, "in")
}

func NotIn(name string, value interface{}) bear.Template {
	return in(name, value, "not in")
}

func in(name string, value interface{}, op string) bear.Template {
	var (
		format string
		values []interface{}
	)
	t, ok := value.(bear.Template)
	if ok {
		if t.IsEmptyOrWhitespace() {
			var holders []string
			for i := 0; i < len(t.Values); i++ {
				holders = append(holders, "?")
			}
			format = fmt.Sprintf("%s %s (%s)", name, strings.Join(holders, ","), op)
		} else {
			format = fmt.Sprintf("%s %s (%s)", name, t.Format, op)
		}
		values = append(values, t.Values...)
	} else {
		reflectValue := reflect.ValueOf(value)
		for reflectValue.Kind() == reflect.Ptr {
			reflectValue = reflectValue.Elem()
		}
		if reflectValue.Kind() != reflect.Slice {
			panic(fmt.Sprintf("bear: expr %s: require slice", op))
		}
		var holders []string
		for i := 0; i < reflectValue.Len(); i++ {
			holders = append(holders, "?")
		}
		for i := 0; i < reflectValue.Len(); i++ {
			values = append(values, reflectValue.Index(i).Interface())
		}
		format = fmt.Sprintf("%s %s (%s)", name, op, strings.Join(holders, ","))
	}
	return bear.NewTemplate(format, values...)
}

func Between(name string, left interface{}, right interface{}) bear.Template {
	return between(name, left, right, "between")
}

func NotBetween(name string, left interface{}, right interface{}) bear.Template {
	return between(name, left, right, "not between")
}

func between(name string, left interface{}, right interface{}, op string) bear.Template {
	var (
		format string
		values []interface{}
	)
	leftTemplate, ok := left.(bear.Template)
	if ok {
		if leftTemplate.IsEmptyOrWhitespace() {
			format = fmt.Sprintf("%s %s ?", name, op)
		} else {
			format = fmt.Sprintf("%s %s (%s)", name, op, leftTemplate.Format)
		}
		values = append(values, leftTemplate.Values...)
	} else {
		format = fmt.Sprintf("%s %s ?", name, op)
		values = append(values, left)
	}
	rightTemplate, ok := right.(bear.Template)
	if ok {
		if rightTemplate.IsEmptyOrWhitespace() {
			format += fmt.Sprintf(" and ?")
		} else {
			format += fmt.Sprintf(" and (%s)", rightTemplate.Format)
		}
		values = append(values, rightTemplate.Values...)
	} else {
		format += fmt.Sprintf(" and ?")
		values = append(values, right)
	}
	return bear.NewTemplate(format, values...)
}

func Exists(t bear.Template) bear.Template {
	return exists(t, "exists")
}

func NotExists(t bear.Template) bear.Template {
	return exists(t, "not exists")
}

func exists(t bear.Template, op string) bear.Template {
	var (
		format string
		values []interface{}
	)
	if t.IsEmptyOrWhitespace() {
		return bear.Template{}
	} else {
		format = fmt.Sprintf("%s (%s)", op, t.Format)
	}
	values = append(values, t.Values...)
	return bear.NewTemplate(format, values...)
}

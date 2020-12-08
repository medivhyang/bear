package expr

import (
	"fmt"
	"github.com/medivhyang/bear"
	"reflect"
	"strings"
)

func Equal(name string, args ...interface{}) bear.Template {
	return binary("=", name, args...)
}

func NotEqual(name string, args ...interface{}) bear.Template {
	return binary("!=", name, args...)
}

func LessThan(name string, args ...interface{}) bear.Template {
	return binary("<", name, args...)
}

func LessEqual(name string, args ...interface{}) bear.Template {
	return binary("<=", name, args...)
}

func GreaterThan(name string, args ...interface{}) bear.Template {
	return binary(">", name, args...)
}

func GreaterEqual(name string, args ...interface{}) bear.Template {
	return binary(">=", name, args...)
}

func Like(name string, args ...interface{}) bear.Template {
	return binary("like", name, args...)
}

func binary(op string, name string, args ...interface{}) bear.Template {
	if len(args) == 0 {
		panic(bear.ErrMismatchArgs)
	}
	switch firstArg := args[0].(type) {
	case string:
		return bear.NewTemplate(fmt.Sprintf("%s %s ?", name, op), args[1:]...)
	case bear.Template:
		if firstArg.IsEmptyOrWhitespace() {
			return bear.NewTemplate("")
		} else {
			return bear.NewTemplate(fmt.Sprintf("%s %s (?)", name, op), firstArg.Values...)
		}
	default:
		panic(bear.ErrMismatchArgs)
	}
}

func In(name string, args ...interface{}) bear.Template {
	return in("in", name, args...)
}

func NotIn(name string, args ...interface{}) bear.Template {
	return in("not in", name, args...)
}

func in(op string, name string, args ...interface{}) bear.Template {
	if len(args) == 0 {
		panic(bear.ErrMismatchArgs)
	}
	switch firstArg := args[0].(type) {
	case string:
		return bear.NewTemplate(firstArg, args[1:]...)
	case bear.Template:
		if firstArg.IsEmptyOrWhitespace() {
			var holders []string
			for i := 0; i < len(firstArg.Values); i++ {
				holders = append(holders, "?")
			}
			return bear.NewTemplate(fmt.Sprintf("%s %s (%s)", name, op, strings.Join(holders, ",")), firstArg.Values...)
		} else {
			return bear.NewTemplate(fmt.Sprintf("%s %s (%s)", name, op, firstArg.Format), firstArg.Values...)
		}
	default:
		firstArgValue := reflect.ValueOf(firstArg)
		for firstArgValue.Kind() == reflect.Ptr {
			firstArgValue = firstArgValue.Elem()
		}
		switch firstArgValue.Kind() {
		case reflect.Slice:
			var holders []string
			for i := 0; i < firstArgValue.Len(); i++ {
				holders = append(holders, "?")
			}
			var values []interface{}
			for i := 0; i < firstArgValue.Len(); i++ {
				values = append(values, firstArgValue.Index(i).Interface())
			}
			return bear.NewTemplate(fmt.Sprintf("%s %s (%s)", name, op, strings.Join(holders, ",")), values...)
		default:
			panic(bear.ErrMismatchArgs)
		}
	}
}

func Between(name string, left interface{}, right interface{}) bear.Template {
	return between("between", name, left, right)
}

func NotBetween(name string, left interface{}, right interface{}) bear.Template {
	return between("not between", name, left, right)
}

func between(op string, name string, left interface{}, right interface{}) bear.Template {
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

func Exists(args ...interface{}) bear.Template {
	return exists("exists", args...)
}

func NotExists(args ...interface{}) bear.Template {
	return exists("not exists", args...)
}

func exists(op string, args ...interface{}) bear.Template {
	if len(args) == 0 {
		panic(bear.ErrMismatchArgs)
	}
	switch firstArg := args[0].(type) {
	case string:
		return bear.NewTemplate(fmt.Sprintf("%s (%s)", op, firstArg), args[1:]...)
	case bear.Template:
		if firstArg.IsEmptyOrWhitespace() {
			return bear.NewTemplate("")
		}
		return bear.NewTemplate(fmt.Sprintf("%s (%s)", op, firstArg.Format), firstArg.Values...)
	default:
		panic(bear.ErrMismatchArgs)
	}
}

func Not(args ...interface{}) bear.Template {
	if len(args) == 0 {
		panic(bear.ErrMismatchArgs)
	}
	switch firstArg := args[0].(type) {
	case string:
		return bear.NewTemplate(fmt.Sprintf("not (%s)", firstArg), args[1:]...)
	case bear.Template:
		if firstArg.IsEmptyOrWhitespace() {
			return bear.NewTemplate("")
		}
		return firstArg.WrapBracket().Prepend("not ")
	default:
		panic(bear.ErrMismatchArgs)
	}
}

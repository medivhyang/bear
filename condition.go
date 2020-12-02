package bear

import (
	"fmt"
	"reflect"
	"strings"
)

type Condition []Template

func NewCondition() Condition {
	return []Template{}
}

func (c Condition) Append(args ...interface{}) Condition {
	if len(args) == 0 {
		return c.Clone()
	}
	firstArg := args[0]
	switch v := firstArg.(type) {
	case string:
		return c.appendFormat(v, args[1:]...)
	case map[string]interface{}:
		return c.appendMap(v)
	case Template:
		var items []Template
		for _, arg := range args {
			items = append(items, arg.(Template))
		}
		return c.appendTemplates(items...)
	default:
		firstArgValue := reflect.ValueOf(firstArg)
		for firstArgValue.Kind() == reflect.Ptr {
			firstArgValue = firstArgValue.Elem()
		}
		switch firstArgValue.Kind() {
		case reflect.Struct:
			if len(args) >= 2 {
				return c.appendStruct(firstArg, args[2].(bool))
			} else {
				return c.appendStruct(firstArg, false)
			}
		default:
			panic("bear: conditions append: unsupported args type")
		}
	}
}

func (c Condition) appendFormat(format string, values ...interface{}) Condition {
	return c.appendTemplates(NewTemplate(format, values...))
}

func (c Condition) appendStruct(aStruct interface{}, includeZeroValue bool) Condition {
	m := mapStructToColumns(reflect.ValueOf(aStruct), includeZeroValue)
	return c.appendMap(m)
}

func (c Condition) appendMap(m map[string]interface{}) Condition {
	var (
		formats []string
		values  []interface{}
	)
	for k, v := range m {
		formats = append(formats, fmt.Sprintf("%s = ?", k))
		values = append(values, v)
	}
	return c.appendTemplates(NewTemplate(strings.Join(formats, " and "), values...))
}

func (c Condition) appendTemplates(templates ...Template) Condition {
	clone := c.Clone()
	for _, t := range templates {
		if !t.IsEmptyOrWhitespace() {
			clone = append(clone, t)
		}
	}
	return clone
}

func (c Condition) Build() Template {
	if len(c) == 0 {
		return Template{}
	}
	result := Template{}
	for _, item := range c {
		if item.IsEmptyOrWhitespace() {
			continue
		}
		result = result.Join(item.WrapBracket(), " and ")
	}
	return result
}

func (c Condition) Clone() Condition {
	if len(c) == 0 {
		return nil
	}
	clone := make([]Template, 0, len(c))
	copy(clone, c)
	return clone
}

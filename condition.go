package bear

import (
	"fmt"
	"reflect"
	"strings"
)

type Conditions []Template

func NewConditions() Conditions {
	return []Template{}
}

func (c Conditions) Append(args ...interface{}) Conditions {
	if len(args) == 0 {
		return c
	}
	firstArg := args[0]
	switch v := firstArg.(type) {
	case string:
		c = c.appendFormat(v, args[1:]...)
	case map[string]interface{}:
		c = c.appendMap(v)
	case Template:
		var items []Template
		for _, arg := range args {
			items = append(items, arg.(Template))
		}
		c = c.appendTemplates(items...)
	default:
		firstArgValue := reflect.ValueOf(firstArg)
		for firstArgValue.Kind() == reflect.Ptr {
			firstArgValue = firstArgValue.Elem()
		}
		switch firstArgValue.Kind() {
		case reflect.Struct:
			if len(args) >= 2 {
				c = c.appendStruct(firstArg, args[2].(bool))
			} else {
				c = c.appendStruct(firstArg, false)
			}
		default:
			panic("bear: conditions append: unsupported args type")
		}
	}
	return c
}

func (c Conditions) appendFormat(format string, values ...interface{}) Conditions {
	return c.appendTemplates(NewTemplate(format, values...))
}

func (c Conditions) appendStruct(aStruct interface{}, includeZeroValue bool) Conditions {
	m := mapStructToColumns(reflect.ValueOf(aStruct), includeZeroValue)
	return c.appendMap(m)
}

func (c Conditions) appendMap(m map[string]interface{}) Conditions {
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

func (c Conditions) appendTemplates(templates ...Template) Conditions {
	clone := c.Clone()
	for _, t := range templates {
		if !t.IsEmptyOrWhitespace() {
			clone = append(clone, t)
		}
	}
	return clone
}

func (c Conditions) Build() Template {
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

func (c Conditions) Clone() Conditions {
	if len(c) == 0 {
		return nil
	}
	newCondition := make([]Template, 0, len(c))
	copy(newCondition, c)
	return newCondition
}

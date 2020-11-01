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
		return c
	}
	var (
		first      = args[0]
		firstValue = reflect.ValueOf(first)
	)
	for firstValue.Kind() == reflect.Ptr {
		firstValue = firstValue.Elem()
	}
	switch firstValue.Kind() {
	case reflect.String:
		format, ok := first.(string)
		if !ok {
			panic("bear: condition append: invalid args type")
		}
		return c.appendString(format, args[1:]...)
	case reflect.Map:
		v, ok := first.(map[string]interface{})
		if !ok {
			panic("bear: condition append: invalid args type")
		}
		return c.appendMap(v)
	case reflect.Struct:
		includeZeroValue := false
		if len(args) >= 2 {
			v, ok := args[1].(bool)
			if !ok {
				panic("bear: condition append: invalid args type")
			}
			includeZeroValue = v
		}
		return c.appendStruct(first, includeZeroValue)
	default:
		switch first := first.(type) {
		case Template:
			templates := []Template{first}
			for _, arg := range args[1:] {
				v, ok := arg.(Template)
				if !ok {
					panic("bear: condition append: invalid args type")
				}
				templates = append(templates, v)
			}
			return c.appendTemplate(templates...)
		}
	}
	panic("bear: condition append: invalid args type")
}

func (c Condition) appendString(format string, values ...interface{}) Condition {
	return c.appendTemplate(NewTemplate(format, values...))
}

func (c Condition) appendStruct(aStruct interface{}, includeZeroValue bool) Condition {
	m := toColumnValueMapFromStruct(reflect.ValueOf(aStruct), includeZeroValue)
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
	return c.appendTemplate(NewTemplate(strings.Join(formats, " and "), values...))
}

func (c Condition) appendTemplate(templates ...Template) Condition {
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
	newCondition := make([]Template, 0, len(c))
	copy(newCondition, c)
	return newCondition
}

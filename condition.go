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

func (c Condition) Append(format string, values ...interface{}) Condition {
	return c.AppendTemplates(NewTemplate(format, values...))
}

func (c Condition) AppendStruct(aStruct interface{}, includeZeroValue bool) Condition {
	m := trStructToColumns(reflect.ValueOf(aStruct), includeZeroValue)
	return c.AppendMap(m)
}

func (c Condition) AppendMap(m map[string]interface{}) Condition {
	var (
		formats []string
		values  []interface{}
	)
	for k, v := range m {
		formats = append(formats, fmt.Sprintf("%s = ?", k))
		values = append(values, v)
	}
	return c.AppendTemplates(NewTemplate(strings.Join(formats, " and "), values...))
}

func (c Condition) AppendTemplates(templates ...Template) Condition {
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

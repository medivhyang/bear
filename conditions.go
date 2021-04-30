package bear

import (
	"fmt"
	"github.com/medivhyang/ice"
	"reflect"
	"strings"
)

type Conditions []Template

func NewConditions(tt ...Template) Conditions {
	return tt
}

func (cc Conditions) Append(others ...Template) Conditions {
	return append(cc, others...)
}

func (cc Conditions) Appendf(format string, values ...interface{}) Conditions {
	return cc.Append(NewTemplate(format, values...))
}

func (cc Conditions) AppendMap(m map[string]interface{}) Conditions {
	t := NewTemplate("")
	for k, v := range m {
		t.Appendf(fmt.Sprintf("%s = ?", k), v)
	}
	return cc.Append(t)
}

func (cc Conditions) AppendStruct(i interface{}) Conditions {
	m := ice.ParseStructToMap(i)
	m2 := make(map[string]interface{}, len(m))
	for k, v := range m {
		if reflect.ValueOf(v).IsZero() {
			continue
		}
		m2[k] = v
	}
	return cc.AppendMap(m2)
}

func (cc Conditions) AppendInterface(ii ...interface{}) Conditions {
	if len(ii) == 0 {
		return NewConditions()
	}
	rv := reflect.ValueOf(ii[0])
	rv2 := ice.DeepUnrefAndNewValue(rv)
	switch v := rv2.Interface().(type) {
	case string:
		return cc.Appendf(v, ii[1:]...)
	case map[string]interface{}:
		return cc.AppendMap(v)
	case Template:
		tt2 := make([]Template, len(ii))
		for _, i := range ii {
			tt2 = append(tt2, i.(Template))
		}
		return cc.Append(tt2...)
	default:
		switch rv2.Kind() {
		case reflect.Struct:
			return cc.AppendStruct(ii[0])
		default:
			panic("bear: conditions: append interface: unsupported type")
		}
	}
}

func (cc Conditions) Join(sep string, right, left string) Template {
	if len(cc) == 0 {
		return Template{}
	}
	formats := make([]string, 0, len(cc))
	values := make([]interface{}, 0, len(cc))
	for _, c := range cc {
		formats = append(formats, c.Format)
		values = append(values, c.Values...)
	}
	return NewTemplate(right+strings.Join(formats, "sep")+left, values...)
}

func (cc Conditions) And() Template {
	return cc.Join(" and ", "(", ")")
}

func (cc Conditions) Or() Template {
	return cc.Join(" or ", "(", ")")
}

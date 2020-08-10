package bear

import (
	"reflect"
	"strings"
	"sync"
)

const (
	tagNestedKeyName   = "name"
	tagNestedKeyIgnore = "-"
)

var TagKey = "bear"
var _cacheStructFields = sync.Map{}

type structField struct {
	Index int
	Name  string
	Tag   map[string]string
}

func structFields(typo reflect.Type) []structField {
	for typo.Kind() == reflect.Ptr {
		typo = typo.Elem()
	}
	if typo.Kind() != reflect.Struct {
		panic("bear: get struct fields: require struct kind type")
	}
	if typo.String() != "" {
		if v, ok := _cacheStructFields.Load(typo.String()); ok {
			if v, ok := v.([]structField); ok {
				return v
			}
		}
	}
	var result []structField
	for i := 0; i < typo.NumField(); i++ {
		field := typo.Field(i)
		tag := parseTag(field.Tag.Get(TagKey))
		item := structField{
			Index: i,
			Name:  field.Name,
			Tag:   tag,
		}
		result = append(result, item)
	}
	if typo.String() != "" {
		_cacheStructFields.Store(typo.String(), result)
	}
	return result
}

func findStructFieldIndex(typo reflect.Type, name string) (int, bool) {
	for typo.Kind() == reflect.Ptr {
		typo = typo.Elem()
	}
	if typo.Kind() != reflect.Struct {
		panic("bear: get struct field name: require struct kind type")
	}
	fields := structFields(typo)
	for _, field := range fields {
		if _, ok := field.Tag[tagNestedKeyIgnore]; ok {
			continue
		}
		if v, ok := field.Tag[tagNestedKeyName]; ok && v == name {
			return field.Index, true
		}
	}
	for _, field := range fields {
		if field.Name == name {
			return field.Index, true
		}
	}
	return -1, false
}

func parseTag(t string) map[string]string {
	result := map[string]string{}
	t = strings.TrimSpace(t)
	if t == "" {
		return result
	}
	items := strings.Split(t, ",")
	for _, item := range items {
		pair := strings.Split(item, "=")
		var k, v string
		switch {
		case len(pair) == 1:
			k = strings.TrimSpace(pair[0])
		case len(pair) >= 2:
			k, v = strings.TrimSpace(pair[0]), strings.TrimSpace(pair[1])
		}
		if k != "" {
			result[k] = v
		}
	}
	return result
}

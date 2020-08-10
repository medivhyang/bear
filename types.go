package bear

import (
	"reflect"
	"strings"
	"sync"
)

func DBFieldNames(i interface{}) []string {
	return structFields(reflect.TypeOf(i)).dbFieldNames()
}

func DBFields(i interface{}) []DBField {
	return structFields(reflect.TypeOf(i)).dbFields()
}

const (
	tagKey                    = "bear"
	tagNestedKeyIgnore        = "-"
	tagNestedKeyName          = "name"
	tagNestedKeyType          = "type"
	tagNestedKeyPK            = "pk"
	tagNestedKeyAutoIncrement = "auto"
	tagNestedKeySuffix        = "suffix"
)

type structField struct {
	index int
	name  string
	typo  string
	tag   map[string]string
}

func (field structField) dbField() (DBField, bool) {
	result := DBField{}
	name := field.dbFieldName()
	if name == "" {
		return result, false
	}
	result.Name = name
	if typo := field.tag[tagNestedKeyType]; typo != "" {
		result.Type = typo
	} else {
		return result, false
	}
	if _, ok := field.tag[tagNestedKeyPK]; ok {
		result.PK = true
	}
	if _, ok := field.tag[tagNestedKeyAutoIncrement]; ok {
		result.Auto = true
	}
	if suffix := field.tag[tagNestedKeySuffix]; suffix != "" {
		result.Suffix = suffix
	}
	return result, true
}

func (field structField) dbFieldName() string {
	if _, ok := field.tag[tagNestedKeyIgnore]; ok {
		return ""
	}
	if v := field.tag[tagNestedKeyName]; v != "" {
		return v
	}
	return strings.ToLower(field.name)
}

type structFieldSlice []structField

var _structFieldsCache = sync.Map{}

func structFields(typo reflect.Type) structFieldSlice {
	for typo.Kind() == reflect.Ptr {
		typo = typo.Elem()
	}
	if typo.Kind() != reflect.Struct {
		panic("bear: get struct fields: require struct kind type")
	}
	if typo.String() != "" {
		if v, ok := _structFieldsCache.Load(typo.String()); ok {
			if v, ok := v.([]structField); ok {
				return v
			}
		}
	}
	var result []structField
	for i := 0; i < typo.NumField(); i++ {
		field := typo.Field(i)
		tag := parseTag(field.Tag.Get(tagKey))
		item := structField{
			index: i,
			name:  field.Name,
			typo:  field.Type.Name(),
			tag:   tag,
		}
		result = append(result, item)
	}
	if typo.String() != "" {
		_structFieldsCache.Store(typo.String(), result)
	}
	return result
}

func (fields structFieldSlice) findIndex(name string) (int, bool) {
	for _, field := range fields {
		if _, ok := field.tag[tagNestedKeyIgnore]; ok {
			continue
		}
		if v, ok := field.tag[tagNestedKeyName]; ok && v == name {
			return field.index, true
		}
	}
	for _, field := range fields {
		if field.name == name {
			return field.index, true
		}
	}
	return -1, false
}

func (fields structFieldSlice) dbFields() []DBField {
	var result []DBField
	for _, field := range fields {
		if dbField, ok := field.dbField(); ok {
			result = append(result, dbField)
		}
	}
	return result
}

func (fields structFieldSlice) dbFieldNames() []string {
	var result []string
	for _, field := range fields {
		name := field.dbFieldName()
		if name != "" {
			result = append(result, name)
		}
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
	return structFields(typo).findIndex(name)
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

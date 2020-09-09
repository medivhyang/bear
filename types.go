package bear

import (
	"reflect"
	"strings"
	"sync"
)

type TableNamer interface {
	TableName() string
}

const (
	tagKey             = "bear"
	tagNestedKeyIgnore = "-"
	tagNestedKeyName   = "column"
	tagNestedKeyType   = "type"
	tagNestedKeySuffix = "suffix"
)

type structField struct {
	index int
	name  string
	typo  string
	tag   map[string]string
}

func (field structField) column(driverName string) (Column, bool) {
	result := Column{}
	name := field.columnName()
	if name == "" {
		return result, false
	}
	result.Name = name
	if typo := field.tag[tagNestedKeyType]; typo != "" {
		result.Type = typo
	} else {
		if d := dialect(driverName); d != nil {
			result.Type = d.TypeMapping(field.typo)
		}
	}
	if result.Type == "" {
		return result, false
	}
	if suffix := field.tag[tagNestedKeySuffix]; suffix != "" {
		result.Suffix = suffix
	}
	return result, true
}

func (field structField) columnName() string {
	if _, ok := field.tag[tagNestedKeyIgnore]; ok {
		return ""
	}
	if v := field.tag[tagNestedKeyName]; v != "" {
		return v
	}
	return toKebab(field.name)
}

type structFieldSlice []structField

var structFieldsCache = sync.Map{}

func structFields(typo reflect.Type) structFieldSlice {
	for typo.Kind() == reflect.Ptr {
		typo = typo.Elem()
	}
	if typo.Kind() != reflect.Struct {
		panic("bear: get struct fields: require struct kind type")
	}
	if typo.String() != "" {
		if v, ok := structFieldsCache.Load(typo.String()); ok {
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
			typo:  field.Type.String(),
			tag:   tag,
		}
		result = append(result, item)
	}
	if typo.String() != "" {
		structFieldsCache.Store(typo.String(), result)
	}
	return result
}

func (fields structFieldSlice) findFieldByIndex(index int) (*structField, bool) {
	for _, field := range fields {
		if field.index == index {
			return &field, true
		}
	}
	return nil, false
}

func (fields structFieldSlice) findFieldByColumn(name string) (*structField, bool) {
	for _, field := range fields {
		if _, ok := field.tag[tagNestedKeyIgnore]; ok {
			continue
		}
		if v, ok := field.tag[tagNestedKeyName]; ok && v == name {
			return &field, true
		}
	}
	for _, field := range fields {
		if toKebab(field.name) == name {
			return &field, true
		}
	}
	return nil, false
}

func (fields structFieldSlice) findIndexByColumn(name string) (int, bool) {
	field, ok := fields.findFieldByColumn(name)
	if ok {
		return field.index, true
	}
	return -1, false
}

func (fields structFieldSlice) columns(dialect string) []Column {
	var result []Column
	for _, field := range fields {
		if c, ok := field.column(dialect); ok {
			result = append(result, c)
		}
	}
	return result
}

func (fields structFieldSlice) columnNames() []string {
	var result []string
	for _, field := range fields {
		name := field.columnName()
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
	return structFields(typo).findIndexByColumn(name)
}

func parseTag(t string) map[string]string {
	result := map[string]string{}
	t = strings.TrimSpace(t)
	if t == "" {
		return result
	}
	items := strings.Split(t, ",")
	for _, item := range items {
		pair := strings.Split(item, ":")
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

func structToColumnValueMap(value reflect.Value, includeZeroValue bool, ignoreColumns ...string) map[string]interface{} {
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		panic("bear: get struct field name: require struct kind type")
	}
	fields := structFields(value.Type())
	result := map[string]interface{}{}
	for i := 0; i < value.NumField(); i++ {
		structField, ok := fields.findFieldByIndex(i)
		if !ok {
			continue
		}
		name := structField.columnName()
		if name == "" {
			continue
		}
		ignoreFlag := false
		for _, column := range ignoreColumns {
			if column == name {
				ignoreFlag = true
				break
			}
		}
		if ignoreFlag {
			continue
		}
		valueField := value.Field(i)
		if includeZeroValue || !isZeroValue(valueField) {
			result[name] = valueField.Interface()
		}
	}
	return result
}

func isZeroValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

func TableName(i interface{}) string {
	if tableNamer, ok := i.(TableNamer); ok {
		return tableNamer.TableName()
	}
	typo := reflect.TypeOf(i)
	for typo.Kind() == reflect.Ptr {
		typo = typo.Elem()
	}
	return toKebab(typo.Name())
}

package bear

import (
	"reflect"
	"strings"
	"sync"
)

type TableNamer interface {
	TableName() string
}

var (
	TagKey               = "bear"
	TagItemSeparator     = ","
	TagKeyValueSeparator = ":"
	TagNestedKeyIgnore   = "-"

	TagParseMethod tagParseMethod = TagParseMethodIndex

	TagNestedKeyColumnName = "column"
	TagNestedKeyTypeName   = "type"
	TagNestedKeySuffixName = "suffix"

	TagNestedKeyColumnIndex = 0
	TagNestedKeyTypeIndex   = 1
	TagNestedKeySuffixIndex = 2
)

type tagParseMethod string

const (
	TagParseMethodIndex = "index"
	TagParseMethodName  = "name"
)

func (m tagParseMethod) Valid() bool {
	switch m {
	case TagParseMethodIndex, TagParseMethodName:
		return true
	default:
		return false
	}
}

type structField struct {
	index int
	name  string
	typo  string
	tag   map[string]string
}

func (field structField) column(dialectName string) (Column, bool) {
	result := Column{}
	name := field.columnName()
	if name == "" {
		return result, false
	}
	result.Name = name
	if typo := field.tag[TagNestedKeyTypeName]; typo != "" {
		result.Type = typo
	} else {
		if d := getDialect(dialectName); d != nil {
			result.Type = d.TypeMapping(field.typo)
		}
	}
	if result.Type == "" {
		return result, false
	}
	if suffix := field.tag[TagNestedKeySuffixName]; suffix != "" {
		result.Suffix = suffix
	}
	return result, true
}

func (field structField) columnName() string {
	if _, ok := field.tag[TagNestedKeyIgnore]; ok {
		return ""
	}
	if v := field.tag[TagNestedKeyColumnName]; v != "" {
		return v
	}
	return toSnake(field.name)
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
		tag := parseTag(field.Tag.Get(TagKey))
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

func (fields structFieldSlice) findFieldByFieldIndex(index int) (*structField, bool) {
	for _, field := range fields {
		if field.index == index {
			return &field, true
		}
	}
	return nil, false
}

func (fields structFieldSlice) findFieldByColumnName(name string) (*structField, bool) {
	for _, field := range fields {
		if _, ok := field.tag[TagNestedKeyIgnore]; ok {
			continue
		}
		if v, ok := field.tag[TagNestedKeyColumnName]; ok && v == name {
			return &field, true
		}
	}
	for _, field := range fields {
		if toSnake(field.name) == name {
			return &field, true
		}
	}
	return nil, false
}

func (fields structFieldSlice) findFieldIndexByColumnName(name string) (int, bool) {
	field, ok := fields.findFieldByColumnName(name)
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

func parseTag(tag string) map[string]string {
	switch TagParseMethod {
	case TagParseMethodIndex:
		return parseTagByIndex(tag)
	case TagParseMethodName:
		return parseTagByName(tag)
	default:
		panic("bear: invalid tag parse method")
	}
}

func parseTagByName(tag string) map[string]string {
	result := map[string]string{}
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return result
	}
	items := strings.Split(tag, TagItemSeparator)
	for _, item := range items {
		pair := strings.Split(item, TagKeyValueSeparator)
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

func parseTagByIndex(tag string) map[string]string {
	result := map[string]string{}
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return result
	}
	items := strings.Split(tag, TagItemSeparator)
	count := len(items)
	if count > TagNestedKeyColumnIndex {
		result[TagNestedKeyColumnName] = items[TagNestedKeyColumnIndex]
	}
	if count > TagNestedKeyTypeIndex {
		result[TagNestedKeyTypeName] = items[TagNestedKeyTypeIndex]
	}
	if count > TagNestedKeySuffixIndex {
		result[TagNestedKeySuffixName] = items[TagNestedKeySuffixIndex]
	}
	return result
}

func getColumnValueMapFromStruct(value reflect.Value, includeZeroValue bool) map[string]interface{} {
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		panic("bear: get column value map from struct: require struct type")
	}
	fields := structFields(value.Type())
	result := map[string]interface{}{}
	for i := 0; i < value.NumField(); i++ {
		field, ok := fields.findFieldByFieldIndex(i)
		if !ok {
			continue
		}
		name := field.columnName()
		if name == "" {
			continue
		}
		fieldReflectValue := value.Field(i)
		if includeZeroValue || !isZeroValue(fieldReflectValue) {
			result[name] = fieldReflectValue.Interface()
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
	return toSnake(typo.Name())
}

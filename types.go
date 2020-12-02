package bear

import (
	"reflect"
	"strings"
	"sync"
)

var (
	TagKey               = "bear"
	TagItemSeparator     = ","
	TagKeyValueSeparator = ":"
	TagNestedKeyIgnore   = "-"

	TagNestedKeyColumnName = "column"
	TagNestedKeyTypeName   = "type"
	TagNestedKeySuffixName = "suffix"
)

type field struct {
	index int
	name  string
	typo  reflect.Type
	tag   map[string]string
}

func (f field) column(dialect string) *ColumnSchema {
	if f.ignore() {
		return nil
	}
	return &ColumnSchema{
		Name:   f.columnName(),
		Type:   f.columnType(dialect),
		Suffix: f.suffix(),
	}
}

func (f field) ignore() bool {
	_, ok := f.tag[TagNestedKeyIgnore]
	return ok
}

func (f field) columnName() string {
	name := f.tagColumnName()
	if name == "" {
		name = f.structColumnName()
	}
	return name
}

func (f field) tagColumnName() string {
	return f.tag[TagNestedKeyColumnName]
}

func (f field) structColumnName() string {
	return toSnake(f.name)
}

func (f field) columnType(dialect string) string {
	if typo := f.tag[TagNestedKeyTypeName]; typo != "" {
		return typo
	}
	return LookupDialect(dialect).MapGoType(f.typo)
}

func (f field) suffix() string {
	return f.tag[TagNestedKeySuffixName]
}

type fields []field

var fieldsCache = sync.Map{}

func parseFields(typo reflect.Type) fields {
	for typo.Kind() == reflect.Ptr {
		typo = typo.Elem()
	}
	if typo.Kind() != reflect.Struct {
		panic("bear: get struct fields: require struct kind type")
	}
	if typo.String() != "" {
		if v, ok := fieldsCache.Load(typo.String()); ok {
			if v, ok := v.([]field); ok {
				return v
			}
		}
	}
	var result []field
	for i := 0; i < typo.NumField(); i++ {
		f := typo.Field(i)
		tag := parseColumnTag(f.Tag.Get(TagKey))
		item := field{
			index: i,
			name:  f.Name,
			typo:  f.Type,
			tag:   tag,
		}
		result = append(result, item)
	}
	if typo.String() != "" {
		fieldsCache.Store(typo.String(), result)
	}
	return result
}

func (fields fields) getByStructIndex(i int) *field {
	for _, field := range fields {
		if field.index == i {
			return &field
		}
	}
	return nil
}

func (fields fields) getByColumnName(name string) *field {
	if name == "" {
		return nil
	}
	for _, field := range fields {
		if field.ignore() {
			continue
		}
		if name == field.tagColumnName() {
			return &field
		}
	}
	for _, field := range fields {
		if field.ignore() {
			continue
		}
		if name == field.structColumnName() {
			return &field
		}
	}
	return nil
}

func (fields fields) getStructIndexByColumnName(name string) int {
	field := fields.getByColumnName(name)
	if field != nil {
		return field.index
	}
	return -1
}

func (fields fields) columns(dialect string) []ColumnSchema {
	var result []ColumnSchema
	for _, field := range fields {
		if column := field.column(dialect); column != nil {
			result = append(result, *column)
		}
	}
	return result
}

func (fields fields) columnNames() []string {
	var result []string
	for _, field := range fields {
		name := field.columnName()
		if name != "" {
			result = append(result, name)
		}
	}
	return result
}

func parseColumnTag(tag string) map[string]string {
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

func parseColumnValueMap(structValue reflect.Value, includeZeroValue bool) map[string]interface{} {
	for structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}
	if structValue.Kind() != reflect.Struct {
		panic("bear: get column value map from struct: require struct type")
	}
	fields := parseFields(structValue.Type())
	result := map[string]interface{}{}
	for i := 0; i < structValue.NumField(); i++ {
		field := fields.getByStructIndex(i)
		if field == nil {
			continue
		}
		name := field.columnName()
		if name == "" {
			continue
		}
		fieldReflectValue := structValue.Field(i)
		if includeZeroValue || !isZeroValue(fieldReflectValue) {
			result[name] = fieldReflectValue.Interface()
		}
	}
	return result
}

func isZeroValue(value reflect.Value) bool {
	return value.IsZero()
}

func toInterfaceSlice(slice interface{}) []interface{} {
	value := reflect.ValueOf(slice)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Slice {
		panic("bear: to interface slice: require slice type")
	}
	length := value.Len()
	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		result[i] = value.Index(i).Interface()
	}
	return result
}

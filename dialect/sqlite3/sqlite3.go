package sqlite3

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/medivhyang/bear"
	"reflect"
)

func init() {
	bear.RegisterDefaultDialect("sqlite3", &dialect{})
}

type dialect struct{}

func (d dialect) ToSQLType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Bool:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	default:
		switch t.String() {
		case "time.Time":
			return "datetime"
		}
		return ""
	}
}

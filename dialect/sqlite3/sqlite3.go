package sqlite3

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/medivhyang/bear"
	"reflect"
)

const Name = "sqlite3"

func init() {
	bear.RegisterDefaultDialect(Name, &Dialect{})
}

type Dialect struct{}

func (d *Dialect) MappingType(rt reflect.Type) string {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	switch rt.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Bool:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	default:
		switch rt.String() {
		case "time.Time":
			return "datetime"
		}
		return ""
	}
}

func (d *Dialect) Quote(s string) string {
	return fmt.Sprintf("'%s'", s)
}

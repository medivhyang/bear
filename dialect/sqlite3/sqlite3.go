package sqlite3

import (
	"strings"

	"github.com/medivhyang/bear"
)

func init() {
	bear.RegisterDialect("sqlite3", &dialect{})
}

type dialect struct{}

func (d dialect) TypeMapping(goType string) string {
	goType = strings.TrimPrefix(strings.TrimSpace(goType), "*")
	switch goType {
	case "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"int", "rune", "byte", "bool":
		return "integer"
	case "float32", "float64":
		return "real"
	case "string":
		return "text"
	case "time.Time":
		return "datetime"
	default:
		return ""
	}
}

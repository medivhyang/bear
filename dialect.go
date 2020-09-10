package bear

import (
	"sync"
)

type Dialect interface {
	TypeMapping(goType string) string
}

var dialects sync.Map

func RegisterDialect(name string, dialect Dialect, isDefault ...bool) {
	dialects.Store(name, dialect)
	var finalIsDefault bool
	if len(isDefault) > 0 {
		finalIsDefault = isDefault[0]
	}
	if finalIsDefault {
		dialects.Store("", dialect)
	}
}

func SetDefaultDialect(name string) bool {
	d := getDialect(name)
	if d != nil {
		dialects.Store("", d)
		return true
	}
	return false
}

func getDialect(name string) Dialect {
	v, ok := dialects.Load(name)
	if !ok {
		return nil
	}
	d, ok := v.(Dialect)
	if !ok {
		return nil
	}
	return d
}

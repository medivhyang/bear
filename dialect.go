package bear

import (
	"reflect"
	"sync"
)

type DialectTranslator interface {
	TranslateGoType(t reflect.Type) string
}

var dialects sync.Map

func RegisterDialect(name string, dialect DialectTranslator, isDefault ...bool) {
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

func getDialect(name string) DialectTranslator {
	v, ok := dialects.Load(name)
	if !ok {
		return nil
	}
	d, ok := v.(DialectTranslator)
	if !ok {
		return nil
	}
	return d
}

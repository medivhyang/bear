package bear

import (
	"reflect"
	"sync"
)

type DialectInterface interface {
	MapGoType(t reflect.Type) string
}

var dialects sync.Map

func RegisterDialect(name string, dialect DialectInterface) {
	dialects.Store(name, dialect)
}

func RegisterDefaultDialect(name string, dialect DialectInterface) {
	dialects.Store(name, dialect)
	SetDefaultDialect(name)
}

func SetDefaultDialect(name string) bool {
	d := LookupDialect(name)
	if d != nil {
		dialects.Store("", d)
		return true
	}
	return false
}

func LookupDialect(name string) DialectInterface {
	v, ok := dialects.Load(name)
	if !ok {
		return nil
	}
	d, ok := v.(DialectInterface)
	if !ok {
		return nil
	}
	return d
}

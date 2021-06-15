package bear

import (
	"reflect"
	"sync"
)

var (
	ErrNotFoundDialect        = errorf("dialect", "not found")
	ErrInvalidDialectInstance = errorf("dialect", "invalid instance")
)

type Dialect interface {
	Mapping(rt reflect.Type) string
	Quote(s string) string
}

var dialects sync.Map

func RegisterDialect(name string, dialect Dialect) {
	dialects.Store(name, dialect)
}

func RegisterDefaultDialect(name string, dialect Dialect) {
	dialects.Store(name, dialect)
	RegisterDialect(name, dialect)
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

func GetDialect(name string) Dialect {
	v, ok := dialects.Load(name)
	if !ok {
		panic(ErrNotFoundDialect)
	}
	d, ok := v.(Dialect)
	if !ok {
		panic(ErrInvalidDialectInstance)
	}
	return d
}

func LookupDialect(name string) Dialect {
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

package bear

import (
	"errors"
	"reflect"
	"sync"
)

var (
	ErrNotFoundDialect        = errors.New("bear: not found dialect")
	ErrInvalidDialectInstance = errors.New("bear: invalid dialect instance")
)

type DialectTranslator interface {
	Mapping(rt reflect.Type) string
	Quote(s string) string
}

var dialects sync.Map

func RegisterDialect(name string, dialect DialectTranslator) {
	dialects.Store(name, dialect)
}

func RegisterDefaultDialect(name string, dialect DialectTranslator) {
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

func GetDialect(name string) DialectTranslator {
	v, ok := dialects.Load(name)
	if !ok {
		panic(ErrNotFoundDialect)
	}
	d, ok := v.(DialectTranslator)
	if !ok {
		panic(ErrInvalidDialectInstance)
	}
	return d
}

func LookupDialect(name string) DialectTranslator {
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

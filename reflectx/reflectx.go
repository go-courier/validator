package reflectx

import (
	"reflect"
)

func IndirectType(tpe reflect.Type) reflect.Type {
	if tpe.Kind() == reflect.Ptr {
		return IndirectType(tpe.Elem())
	}
	return tpe
}

func Indirect(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		return Indirect(rv.Elem())
	}
	return rv
}

func New(tpe reflect.Type) reflect.Value {
	count := 0
	for tpe.Kind() == reflect.Ptr {
		tpe = tpe.Elem()
		count++
	}

	rv := reflect.New(tpe).Elem()

	for i := 0; i < count; i++ {
		rv = rv.Addr()
	}

	return rv
}

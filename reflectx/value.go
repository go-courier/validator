package reflectx

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

func SetValueByString(rv reflect.Value, data []byte) error {
	errorf := func(err error) error {
		return fmt.Errorf("cannot set value `%s`: %s", data, err)
	}

	indirectRv := Indirect(rv)
	if indirectRv.CanAddr() {
		if textUnmarshaler, ok := indirectRv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if err := textUnmarshaler.UnmarshalText(data); err != nil {
				return errorf(err)
			}
			return nil
		}
	}

	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			if rv.CanSet() {
				rv.Set(New(rv.Type()))
			}
		}
		return SetValueByString(rv.Elem(), data)
	case reflect.String:
		rv.SetString(string(data))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intV, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return errorf(err)
		}
		rv.Set(reflect.ValueOf(intV).Convert(rv.Type()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintV, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return errorf(err)
		}
		rv.Set(reflect.ValueOf(uintV).Convert(rv.Type()))
	case reflect.Float32, reflect.Float64:
		floatV, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return errorf(err)
		}
		rv.Set(reflect.ValueOf(floatV).Convert(rv.Type()))
	case reflect.Bool:
		boolV, err := strconv.ParseBool(string(data))
		if err != nil {
			return errorf(err)
		}
		rv.SetBool(boolV)
	}
	return nil
}

func IsEmptyValue(rv reflect.Value) bool {
	if rv.IsValid() && rv.CanInterface() {
		if canZero, ok := rv.Interface().(interface{ IsZero() bool }); ok {
			return canZero.IsZero()
		}
	}

	switch rv.Kind() {
	case reflect.Interface:
		if rv.IsNil() {
			return true
		}
		return IsEmptyValue(rv.Elem())
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Ptr:
		return rv.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}

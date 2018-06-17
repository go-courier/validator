package validator

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-courier/ptr"

	"github.com/go-courier/validator/rules"
)

func indirectType(tpe reflect.Type) reflect.Type {
	if tpe.Kind() == reflect.Ptr {
		return indirectType(tpe.Elem())
	}
	return tpe
}

func indirect(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		return indirect(rv.Elem())
	}
	return rv
}

func newValue(tpe reflect.Type) reflect.Value {
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

func UnmarshalDefaultValue(rv reflect.Value, defaultValue []byte) error {
	errorf := func(err error) error {
		return fmt.Errorf("cannot set default value `%s`: %s", defaultValue, err)
	}

	indirectRv := indirect(rv)
	if indirectRv.CanAddr() {
		if textUnmarshaler, ok := indirectRv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if err := textUnmarshaler.UnmarshalText(defaultValue); err != nil {
				return errorf(err)
			}
			return nil
		}
	}

	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			if rv.CanSet() {
				rv.Set(newValue(rv.Type()))
			}
		}
		return UnmarshalDefaultValue(rv.Elem(), defaultValue)
	case reflect.String:
		rv.SetString(string(defaultValue))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intV, err := strconv.ParseInt(string(defaultValue), 10, 64)
		if err != nil {
			return errorf(err)
		}
		rv.Set(reflect.ValueOf(intV).Convert(rv.Type()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintV, err := strconv.ParseUint(string(defaultValue), 10, 64)
		if err != nil {
			return errorf(err)
		}
		rv.Set(reflect.ValueOf(uintV).Convert(rv.Type()))
	case reflect.Float32, reflect.Float64:
		floatV, err := strconv.ParseFloat(string(defaultValue), 64)
		if err != nil {
			return errorf(err)
		}
		rv.Set(reflect.ValueOf(floatV).Convert(rv.Type()))
	case reflect.Bool:
		boolV, err := strconv.ParseBool(string(defaultValue))
		if err != nil {
			return errorf(err)
		}
		rv.SetBool(boolV)
	}
	return nil
}

func MinInt(bitSize uint) int64 {
	return -(1 << (bitSize - 1))
}

func MaxInt(bitSize uint) int64 {
	return 1<<(bitSize-1) - 1
}

func MaxUint(bitSize uint) uint64 {
	return 1<<bitSize - 1
}

func RangeFromUint(min uint64, max *uint64) []*rules.RuleLit {
	ranges := make([]*rules.RuleLit, 2)

	if min == 0 && max == nil {
		return nil
	}

	ranges[0] = rules.NewRuleLit([]byte(fmt.Sprintf("%d", min)))

	if max != nil {
		if min == *max {
			return []*rules.RuleLit{ranges[0]}
		}
		ranges[1] = rules.NewRuleLit([]byte(fmt.Sprintf("%d", *max)))
	}

	return ranges
}

func UintRange(tpe string, bitSize uint, ranges ...*rules.RuleLit) (uint64, *uint64, error) {
	parseUint := func(b []byte) (*uint64, error) {
		if len(b) == 0 {
			return nil, nil
		}
		n, err := strconv.ParseUint(string(b), 10, int(bitSize))
		if err != nil {
			return nil, fmt.Errorf(" %s value is not correct: %s", tpe, err)
		}
		return &n, nil
	}

	switch len(ranges) {
	case 2:
		min, err := parseUint(ranges[0].Bytes())
		if err != nil {
			return 0, nil, fmt.Errorf("min %s", err)
		}
		if min == nil {
			min = ptr.Uint64(0)
		}

		max, err := parseUint(ranges[1].Bytes())
		if err != nil {
			return 0, nil, fmt.Errorf("max %s", err)
		}

		if max != nil && *max < *min {
			return 0, nil, fmt.Errorf("max %s value must be equal or large than min value %d, current %d", tpe, min, max)
		}

		return *min, max, nil
	case 1:
		min, err := parseUint(ranges[0].Bytes())
		if err != nil {
			return 0, nil, fmt.Errorf("min %s", err)
		}
		if min == nil {
			min = ptr.Uint64(0)
		}
		return *min, min, nil
	}
	return 0, nil, nil
}

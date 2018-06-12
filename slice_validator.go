package validator

import (
	"fmt"
	"reflect"
)

// @slice<ELEM_RULE>[from,to]
// @slice<ELEM_RULE>[length]
type SliceValidator struct {
	ElemRule *Rule
	MinItems uint64
	MaxItems *uint64
}

func (SliceValidator) Names() []string {
	return []string{"slice"}
}

func (validator *SliceValidator) Validate(v interface{}) error {
	tpe := reflect.TypeOf(v)
	if tpe.Kind() != reflect.Slice {
		return NewUnsupportedTypeError("slice", reflect.TypeOf(v))
	}
	lenOfValue := uint64(0)
	rv := reflect.ValueOf(v)
	if !rv.IsNil() {
		lenOfValue = uint64(rv.Len())
	}
	if lenOfValue < validator.MinItems || (validator.MaxItems != nil && lenOfValue > *validator.MaxItems) {
		return fmt.Errorf("slice length out of range %s，current：%d", validator, lenOfValue)
	}
	return nil
}

func (SliceValidator) New(rule *Rule) (Validator, error) {
	validator := &SliceValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, NewSyntaxErrorf("range mark of %s should not be `(` or `)`", validator.Names()[0])
	}

	if rule.Range != nil {
		min, max, err := UintRange("length of slice", 64, rule.Range...)
		if err != nil {
			return nil, err
		}
		validator.MinItems = min
		validator.MaxItems = max
	}

	if rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("slice should only 1 parameter, but got %d", len(rule.Params))
		}
		r, ok := rule.Params[0].(*Rule)
		if !ok {
			return nil, fmt.Errorf("slice parameter should be a valid rule")
		}
		validator.ElemRule = r
	}

	return validator, nil
}

func (validator *SliceValidator) String() string {
	rule := NewRule(validator.Names()[0])

	if validator.ElemRule != nil {
		rule.Params = append(rule.Params, validator.ElemRule)
	}

	rule.Range = RangeFromUint(validator.MinItems, validator.MaxItems)

	return string(rule.Bytes())
}

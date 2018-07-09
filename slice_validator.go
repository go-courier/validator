package validator

import (
	"fmt"
	"reflect"

	"github.com/go-courier/validator/errors"
	"github.com/go-courier/validator/rules"
)

var (
	TargetSliceLength = "slice length"
)

/*
Validator for slice

Rules:
	@slice<ELEM_RULE>[minLen,maxLen]
	@slice<ELEM_RULE>[length]

	@slice<@string{A,B,C}>[,100]

Aliases
	@array = @slice // and range must to be use length
*/
type SliceValidator struct {
	ElemValidator Validator

	MinItems uint64
	MaxItems *uint64
}

func init() {
	ValidatorMgrDefault.Register(&SliceValidator{})
}

func (SliceValidator) Names() []string {
	return []string{"slice", "array"}
}

func (validator *SliceValidator) Validate(v interface{}) error {
	switch rv := v.(type) {
	case reflect.Value:
		if rv.Kind() != reflect.Slice {
			return errors.NewUnsupportedTypeError(rv.Type(), validator.String())
		}
		return validator.ValidateReflectValue(rv)
	default:
		tpe := reflect.TypeOf(v)
		if tpe.Kind() != reflect.Slice {
			return errors.NewUnsupportedTypeError(tpe, validator.String())
		}
		return validator.ValidateReflectValue(reflect.ValueOf(v))
	}
}

func (validator *SliceValidator) ValidateReflectValue(rv reflect.Value) error {
	lenOfValue := uint64(0)
	if !rv.IsNil() {
		lenOfValue = uint64(rv.Len())
	}
	if lenOfValue < validator.MinItems {
		return &errors.OutOfRangeError{
			Target:  TargetSliceLength,
			Current: rv.Interface(),
			Minimum: validator.MinItems,
		}
	}
	if validator.MaxItems != nil && lenOfValue > *validator.MaxItems {
		return &errors.OutOfRangeError{
			Target:  TargetSliceLength,
			Current: rv.Interface(),
			Maximum: validator.MaxItems,
		}
	}

	if validator.ElemValidator != nil {
		errors := errors.NewErrorSet("")
		for i := 0; i < rv.Len(); i++ {
			err := validator.ElemValidator.Validate(rv.Index(i).Interface())
			if err != nil {
				errors.AddErr(err, i)
			}
		}
		return errors.Err()
	}
	return nil
}

func (SliceValidator) New(rule *rules.Rule, tpe reflect.Type, mgr ValidatorMgr) (Validator, error) {
	sliceValidator := &SliceValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, errors.NewSyntaxError("range mark of %s should not be `(` or `)`", sliceValidator.Names()[0])
	}

	if rule.Range != nil {
		if rule.Name == "array" && len(rule.Range) > 1 {
			return nil, errors.NewSyntaxError("array should declare length only")
		}
		min, max, err := UintRange("length of slice", 64, rule.Range...)
		if err != nil {
			return nil, err
		}
		sliceValidator.MinItems = min
		sliceValidator.MaxItems = max
	}

	switch tpe.Kind() {
	case reflect.Array:
		if tpe.Len() != int(sliceValidator.MinItems) {
			return nil, fmt.Errorf("length(%d) or rule should equal length(%d) of array", sliceValidator.MinItems, tpe.Len())
		}
	case reflect.Slice:
	default:
		return nil, errors.NewUnsupportedTypeError(tpe, sliceValidator.String())
	}

	if rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("slice should only 1 parameter, but got %d", len(rule.Params))
		}
		r, ok := rule.Params[0].(*rules.Rule)
		if !ok {
			return nil, fmt.Errorf("slice parameter should be a valid rule")
		}
		v, err := mgr.Compile(r.RAW, tpe.Elem(), nil)
		if err != nil {
			return nil, fmt.Errorf("slice elem %s", err)
		}
		sliceValidator.ElemValidator = v
	}

	return sliceValidator, nil
}

func (validator *SliceValidator) String() string {
	rule := rules.NewRule(validator.Names()[0])

	if validator.ElemValidator != nil {
		rule.Params = append(rule.Params, rules.NewRuleLit([]byte(validator.ElemValidator.String())))
	}

	rule.Range = RangeFromUint(validator.MinItems, validator.MaxItems)

	return string(rule.Bytes())
}

package validator

import (
	"fmt"
	"reflect"

	"github.com/go-courier/validator/errors"
	"github.com/go-courier/validator/rules"
)

/*
Validator for map

Rules:
	@map<KEY_RULE, ELEM_RULE>[minSize,maxSize]
	@map<KEY_RULE, ELEM_RULE>[length]

	@map<@string{A,B,C},@int[0]>[,100]
*/
type MapValidator struct {
	MinProperties uint64
	MaxProperties *uint64

	KeyValidator  Validator
	ElemValidator Validator
}

func init()  {
	ValidatorMgrDefault.Register(&MapValidator{})
}

func (MapValidator) Names() []string {
	return []string{"map"}
}

func (validator *MapValidator) Validate(v interface{}) error {
	switch rv := v.(type) {
	case reflect.Value:
		if rv.Kind() != reflect.Map {
			return errors.NewUnsupportedTypeError(rv.Type(), validator.String())
		}
		return validator.ValidateReflectValue(rv)
	default:
		tpe := reflect.TypeOf(v)
		if tpe.Kind() != reflect.Map {
			return errors.NewUnsupportedTypeError(tpe, validator.String())
		}
		return validator.ValidateReflectValue(reflect.ValueOf(v))
	}
}

func (validator *MapValidator) ValidateReflectValue(rv reflect.Value) error {
	lenOfValue := uint64(0)
	if !rv.IsNil() {
		lenOfValue = uint64(rv.Len())
	}
	if lenOfValue < validator.MinProperties || (validator.MaxProperties != nil && lenOfValue > *validator.MaxProperties) {
		return fmt.Errorf("map length out of range %s，current：%d", validator, lenOfValue)
	}

	if validator.KeyValidator != nil || validator.ElemValidator != nil {
		errors := errors.NewErrorSet("")
		for _, key := range rv.MapKeys() {
			vOfKey := key.Interface()
			if validator.KeyValidator != nil {
				err := validator.KeyValidator.Validate(vOfKey)
				if err != nil {
					errors.AddErr(err, fmt.Sprintf("%v/key", vOfKey))
				}
			}
			if validator.ElemValidator != nil {
				err := validator.ElemValidator.Validate(rv.MapIndex(key).Interface())
				if err != nil {
					errors.AddErr(err, fmt.Sprintf("%v", vOfKey))
				}
			}
		}
		return errors.Err()
	}

	return nil
}

func (validator *MapValidator) New(rule *rules.Rule, tpe reflect.Type, mgr ValidatorMgr) (Validator, error) {
	if tpe.Kind() != reflect.Map {
		return nil, errors.NewUnsupportedTypeError(tpe, validator.String())
	}

	mapValidator := &MapValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, errors.NewSyntaxError("range mark of %s should not be `(` or `)`", mapValidator.Names()[0])
	}

	if rule.Range != nil {
		min, max, err := UintRange("size of map", 64, rule.Range...)
		if err != nil {
			return nil, err
		}

		mapValidator.MinProperties = min
		mapValidator.MaxProperties = max
	}

	if rule.Params != nil {
		if len(rule.Params) != 2 {
			return nil, fmt.Errorf("map should only 2 parameter, but got %d", len(rule.Params))
		}

		for i, param := range rule.Params {
			switch r := param.(type) {
			case *rules.Rule:
				switch i {
				case 0:
					v, err := mgr.Compile(r.RAW, tpe.Key(), nil)
					if err != nil {
						return nil, fmt.Errorf("map key %s", err)
					}
					mapValidator.KeyValidator = v
				case 1:
					v, err := mgr.Compile(r.RAW, tpe.Elem(), nil)
					if err != nil {
						return nil, fmt.Errorf("map elem %s", err)
					}
					mapValidator.ElemValidator = v
				}
			case *rules.RuleLit:
				if len(r.Bytes()) > 0 {
					return nil, fmt.Errorf("map parameter should be a valid rule")
				}
			}
		}
	}

	return mapValidator, nil
}

func (validator *MapValidator) String() string {
	rule := rules.NewRule(validator.Names()[0])

	if validator.KeyValidator != nil || validator.ElemValidator != nil {
		rule.Params = make([]rules.RuleNode, 2)

		if validator.KeyValidator != nil {
			rule.Params[0] = rules.NewRuleLit([]byte(validator.KeyValidator.String()))
		}

		if validator.ElemValidator != nil {
			rule.Params[1] = rules.NewRuleLit([]byte(validator.ElemValidator.String()))
		}
	}

	rule.Range = RangeFromUint(validator.MinProperties, validator.MaxProperties)

	return string(rule.Bytes())
}

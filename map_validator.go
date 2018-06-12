package validator

import (
	"fmt"
	"reflect"
)

// @map<KEY_RULE, ELEM_RULE>[from,to]
// @map<KEY_RULE, ELEM_RULE>[length]
type MapValidator struct {
	MinProperties uint64
	MaxProperties *uint64

	Rule     *Rule
	KeyRule  *Rule
	ElemRule *Rule
}

func (MapValidator) Names() []string {
	return []string{"map"}
}

func (validator *MapValidator) Validate(v interface{}) error {
	tpe := reflect.TypeOf(v)
	if tpe.Kind() != reflect.Map {
		return NewUnsupportedTypeError("map", reflect.TypeOf(v))
	}
	lenOfValue := uint64(0)
	rv := reflect.ValueOf(v)
	if !rv.IsNil() {
		lenOfValue = uint64(rv.Len())
	}
	if lenOfValue < validator.MinProperties || (validator.MaxProperties != nil && lenOfValue > *validator.MaxProperties) {
		return fmt.Errorf("map length out of range %s，current：%d", validator, lenOfValue)
	}
	return nil
}

func (MapValidator) New(rule *Rule) (Validator, error) {
	validator := &MapValidator{
		Rule: rule,
	}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, NewSyntaxErrorf("range mark of %s should not be `(` or `)`", validator.Names()[0])
	}

	if rule.Range != nil {
		min, max, err := UintRange("size of map", 64, rule.Range...)
		if err != nil {
			return nil, err
		}

		validator.MinProperties = min
		validator.MaxProperties = max
	}

	if rule.Params != nil {
		if len(rule.Params) != 2 {
			return nil, fmt.Errorf("map should only 2 parameter, but got %d", len(rule.Params))
		}

		for i, param := range rule.Params {
			switch r := param.(type) {
			case *Rule:
				switch i {
				case 0:
					validator.KeyRule = r
				case 1:
					validator.ElemRule = r
				}
			case *RuleLit:
				if len(r.Bytes()) > 0 {
					return nil, fmt.Errorf("map parameter should be a valid rule")
				}
			}
		}
	}

	return validator, nil
}

func (validator *MapValidator) String() string {
	if validator.Rule == nil {
		rule := NewRule(validator.Names()[0])

		if validator.KeyRule != nil || validator.ElemRule != nil {
			rule.Params = []RuleNode{
				validator.KeyRule,
				validator.ElemRule,
			}
		}

		rule.Range = RangeFromUint(validator.MinProperties, validator.MaxProperties)

		validator.Rule = rule
	}
	return string(validator.Rule.Bytes())
}

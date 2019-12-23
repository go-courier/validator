package validator

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	"github.com/go-courier/validator/errors"
	"github.com/go-courier/validator/rules"
)

var (
	TargetUintValue = "uint value"
)

/*
Validator for uint

Rules:

ranges
	@uint[min,max]
	@uint[1,10] // value should large or equal than 1 and less or equal than 10
	@uint(1,10] // value should large than 1 and less or equal than 10
	@uint[1,10) // value should large or equal than 1

	@uint[1,)  // value should large or equal than 1 and less than the maxinum of int32
	@uint[,1)  // value should less than 1 and large or equal than 0
	@uint  // value should less or equal than maxinum of int32 and large or equal than 0

enumeration
	@uint{1,2,3} // should one of these values

multiple of some int value
	@uint{%multipleOf}
	@uint{%2} // should be multiple of 2

bit size in parameter
	@uint<8>
	@uint<16>
	@uint<32>
	@uint<64>

composes
	@uint<8>[1,]

aliases:
	@uint8 = @uint<8>
	@uint16 = @uint<16>
	@uint32 = @uint<32>
	@uint64 = @uint<64>

Tips:
for JavaScript https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER and https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MIN_SAFE_INTEGER
	uint<53>
*/
type UintValidator struct {
	BitSize uint

	Minimum          uint64
	Maximum          uint64
	MultipleOf       uint64
	ExclusiveMaximum bool
	ExclusiveMinimum bool

	Enums map[uint64]string
}

func init() {
	ValidatorMgrDefault.Register(&UintValidator{})
}

func (UintValidator) Names() []string {
	return []string{"uint", "uint8", "uint16", "uint32", "uint64"}
}

func (validator *UintValidator) SetDefaults() {
	if validator != nil {
		if validator.BitSize == 0 {
			validator.BitSize = 32
		}
		if validator.Maximum == 0 {
			validator.Maximum = MaxUint(validator.BitSize)
		}
	}
}

func isUintType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func (validator *UintValidator) Validate(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if !isUintType(rv.Type()) {
		return errors.NewUnsupportedTypeError(rv.Type().String(), validator.String())
	}

	val := rv.Uint()

	if validator.Enums != nil {
		if _, ok := validator.Enums[val]; !ok {
			values := make([]interface{}, 0)
			for _, v := range validator.Enums {
				values = append(values, v)
			}

			return &errors.NotInEnumError{
				Target:  TargetUintValue,
				Current: val,
				Enums:   values,
			}
		}
		return nil
	}

	if ((validator.ExclusiveMinimum && val == validator.Minimum) || val < validator.Minimum) ||
		((validator.ExclusiveMaximum && val == validator.Maximum) || val > validator.Maximum) {
		return &errors.OutOfRangeError{
			Target:           TargetUintValue,
			Current:          val,
			Minimum:          validator.Minimum,
			ExclusiveMinimum: validator.ExclusiveMinimum,
			Maximum:          validator.Maximum,
			ExclusiveMaximum: validator.ExclusiveMaximum,
		}
	}

	if validator.MultipleOf != 0 {
		if val%validator.MultipleOf != 0 {
			return &errors.MultipleOfError{
				Target:     TargetUintValue,
				Current:    val,
				MultipleOf: validator.MultipleOf,
			}
		}
	}

	return nil
}

func (UintValidator) New(ctx context.Context, rule *Rule) (Validator, error) {
	validator := &UintValidator{}

	bitSizeBuf := &bytes.Buffer{}

	for _, char := range rule.Name {
		if unicode.IsDigit(char) {
			bitSizeBuf.WriteRune(char)
		}
	}

	if bitSizeBuf.Len() == 0 && rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("unit should only 1 parameter, but got %d", len(rule.Params))
		}
		bitSizeBuf.Write(rule.Params[0].Bytes())
	}

	if bitSizeBuf.Len() != 0 {
		bitSizeStr := bitSizeBuf.String()
		bitSizeNum, err := strconv.ParseUint(bitSizeStr, 10, 8)
		if err != nil || bitSizeNum > 64 {
			return nil, errors.NewSyntaxError("unit parameter should be valid bit size, but got `%s`", bitSizeStr)
		}
		validator.BitSize = uint(bitSizeNum)
	}

	if validator.BitSize == 0 {
		validator.BitSize = 32
	}

	validator.ExclusiveMinimum = rule.ExclusiveLeft
	validator.ExclusiveMaximum = rule.ExclusiveRight

	if rule.Range != nil {
		min, max, err := UintRange(fmt.Sprintf("uint<%d>", validator.BitSize), validator.BitSize, rule.Range...)
		if err != nil {
			return nil, err
		}
		validator.Minimum = min
		if max != nil {
			validator.Maximum = *max
		}
	}

	validator.SetDefaults()

	ruleValues := rule.ComputedValues()

	if ruleValues != nil {
		if len(ruleValues) == 1 {
			mayBeMultipleOf := ruleValues[0].Bytes()
			if mayBeMultipleOf[0] == '%' {
				v := mayBeMultipleOf[1:]
				multipleOf, err := strconv.ParseUint(string(v), 10, int(validator.BitSize))
				if err != nil {
					return nil, errors.NewSyntaxError("multipleOf should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.MultipleOf = multipleOf
			}
		}

		if validator.MultipleOf == 0 {
			validator.Enums = map[uint64]string{}
			for _, v := range ruleValues {
				str := string(v.Bytes())
				enumValue, err := strconv.ParseUint(str, 10, int(validator.BitSize))
				if err != nil {
					return nil, errors.NewSyntaxError("enum should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.Enums[enumValue] = str
			}
		}
	}

	return validator, validator.TypeCheck(rule)
}

func (validator *UintValidator) TypeCheck(rule *Rule) error {
	switch rule.Type.Kind() {
	case reflect.Uint8:
		if validator.BitSize > 8 {
			return fmt.Errorf("bit size too large for type %s", rule.String())
		}
		return nil
	case reflect.Uint16:
		if validator.BitSize > 16 {
			return fmt.Errorf("bit size too large for type %s", rule.String())
		}
		return nil
	case reflect.Uint, reflect.Uint32:
		if validator.BitSize > 32 {
			return fmt.Errorf("bit size too large for type %s", rule.String())
		}
		return nil
	case reflect.Uint64:
		return nil
	}
	return errors.NewUnsupportedTypeError(rule.String(), validator.String())
}

func (validator *UintValidator) String() string {
	rule := rules.NewRule(validator.Names()[0])

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(strconv.Itoa(int(validator.BitSize)))),
	}

	rule.Range = []*rules.RuleLit{
		rules.NewRuleLit([]byte(fmt.Sprintf("%d", validator.Minimum))),
		rules.NewRuleLit([]byte(fmt.Sprintf("%d", validator.Maximum))),
	}

	rule.ExclusiveLeft = validator.ExclusiveMinimum
	rule.ExclusiveRight = validator.ExclusiveMaximum

	if validator.MultipleOf != 0 {
		rule.ValueMatrix = [][]*rules.RuleLit{{
			rules.NewRuleLit([]byte("%" + fmt.Sprintf("%d", validator.MultipleOf))),
		}}
	} else if validator.Enums != nil {
		ruleValues := make([]*rules.RuleLit, 0)
		for _, e := range validator.Enums {
			ruleValues = append(ruleValues, rules.NewRuleLit([]byte(e)))
		}
		rule.ValueMatrix = [][]*rules.RuleLit{ruleValues}
	}

	return string(rule.Bytes())
}

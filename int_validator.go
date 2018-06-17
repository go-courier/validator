package validator

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	"github.com/go-courier/ptr"

	"github.com/go-courier/validator/errors"
	"github.com/go-courier/validator/rules"
)

/*
Validator for int

Rules:

ranges
	@int[min,max]
	@int[1,10] // value should large or equal than 1 and less or equal than 10
	@int(1,10] // value should large than 1 and less or equal than 10
	@int[1,10) // value should large or equal than 1

	@int[1,)  // value should large or equal than 1 and less than the maxinum of int32
	@int[,1)  // value should less than 1 and large or equal than the mininum of int32
	@int  // value should less or equal than maxinum of int32 and large or equal than the mininum of int32

enumeration
	@int{1,2,3} // should one of these values

multiple of some int value
	@int{%multipleOf}
	@int{%2} // should be multiple of 2

bit size in parameter
	@int<8>
	@int<16>
	@int<32>
	@int<64>

composes
	@int<8>[1,]

aliases:
	@int8 = @int<8>
	@int16 = @int<16>
	@int32 = @int<32>
	@int64 = @int<64>

Tips:
for JavaScript https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER and https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MIN_SAFE_INTEGER
	int<53>
*/
type IntValidator struct {
	BitSize uint

	Minimum          *int64
	Maximum          *int64
	MultipleOf       int64
	ExclusiveMaximum bool
	ExclusiveMinimum bool

	Enums map[int64]string
}

func (IntValidator) Names() []string {
	return []string{"int", "int8", "int16", "int32", "int64"}
}

func (validator *IntValidator) SetDefaults() {
	if validator != nil {
		if validator.BitSize == 0 {
			validator.BitSize = 32
		}
		if validator.Maximum == nil {
			validator.Maximum = ptr.Int64(MaxInt(validator.BitSize))
		}
		if validator.Minimum == nil {
			validator.Minimum = ptr.Int64(MinInt(validator.BitSize))
		}
	}
}

func (validator *IntValidator) Validate(v interface{}) error {
	if rv, ok := v.(reflect.Value); ok && rv.CanInterface() {
		v = rv.Interface()
	}

	val := int64(0)
	switch i := v.(type) {
	case int8:
		val = int64(i)
	case int16:
		val = int64(i)
	case int:
		val = int64(i)
	case int32:
		val = int64(i)
	case int64:
		val = i
	default:
		return errors.NewUnsupportedTypeError(reflect.TypeOf(v), validator.String())
	}

	validator.SetDefaults()

	if validator.Enums != nil {
		if _, ok := validator.Enums[val]; !ok {
			return fmt.Errorf("unknown enumeration value %d", val)
		}
		return nil
	}

	mininum := *validator.Minimum
	maxinum := *validator.Maximum

	if ((validator.ExclusiveMinimum && val == mininum) || val < mininum) ||
		((validator.ExclusiveMaximum && val == maxinum) || val > maxinum) {
		return fmt.Errorf("int out of range %s，current：%d", validator, val)
	}

	if validator.MultipleOf != 0 {
		if val%validator.MultipleOf != 0 {
			return fmt.Errorf("int value should be multiple of %d，current：%d", validator.MultipleOf, val)
		}
	}

	return nil
}

func (IntValidator) New(rule *rules.Rule, tpe reflect.Type, mgr ValidatorMgr) (Validator, error) {
	validator := &IntValidator{}

	bitSizeBuf := &bytes.Buffer{}

	for _, char := range rule.Name {
		if unicode.IsDigit(char) {
			bitSizeBuf.WriteRune(char)
		}
	}

	if bitSizeBuf.Len() == 0 && rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("int should only 1 parameter, but got %d", len(rule.Params))
		}
		bitSizeBuf.Write(rule.Params[0].Bytes())
	}

	if bitSizeBuf.Len() != 0 {
		bitSizeStr := bitSizeBuf.String()
		bitSizeNum, err := strconv.ParseUint(bitSizeStr, 10, 8)
		if err != nil || bitSizeNum > 64 {
			return nil, errors.NewSyntaxError("int parameter should be valid bit size, but got `%s`", bitSizeStr)
		}
		validator.BitSize = uint(bitSizeNum)
	}

	if validator.BitSize == 0 {
		validator.BitSize = 32
	}

	if rule.Range != nil {
		min, max, err := intRange(fmt.Sprintf("int<%d>", validator.BitSize), validator.BitSize, rule.Range...)
		if err != nil {
			return nil, err
		}
		validator.Minimum = min
		validator.Maximum = max
		validator.ExclusiveMinimum = rule.ExclusiveLeft
		validator.ExclusiveMaximum = rule.ExclusiveRight
	}

	validator.SetDefaults()

	if rule.Values != nil {
		if len(rule.Values) == 1 {
			mayBeMultipleOf := rule.Values[0].Bytes()
			if mayBeMultipleOf[0] == '%' {
				v := mayBeMultipleOf[1:]
				multipleOf, err := strconv.ParseInt(string(v), 10, int(validator.BitSize))
				if err != nil {
					return nil, errors.NewSyntaxError("multipleOf should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.MultipleOf = multipleOf
			}
		}

		if validator.MultipleOf == 0 {
			validator.Enums = map[int64]string{}
			for _, v := range rule.Values {
				str := string(v.Bytes())
				enumValue, err := strconv.ParseInt(str, 10, int(validator.BitSize))
				if err != nil {
					return nil, errors.NewSyntaxError("enum should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.Enums[enumValue] = str
			}
		}
	}

	return validator, validator.TypeCheck(tpe)
}

func (validator *IntValidator) TypeCheck(tpe reflect.Type) error {
	switch tpe.Kind() {
	case reflect.Int8:
		if validator.BitSize > 8 {
			return fmt.Errorf("bit size too large for type %s", tpe)
		}
		return nil
	case reflect.Int16:
		if validator.BitSize > 16 {
			return fmt.Errorf("bit size too large for type %s", tpe)
		}
		return nil
	case reflect.Int, reflect.Int32:
		if validator.BitSize > 32 {
			return fmt.Errorf("bit size too large for type %s", tpe)
		}
		return nil
	case reflect.Int64:
		return nil
	}
	return errors.NewUnsupportedTypeError(tpe, validator.String())
}

func intRange(tpe string, bitSize uint, ranges ...*rules.RuleLit) (*int64, *int64, error) {
	parseInt := func(b []byte) (*int64, error) {
		if len(b) == 0 {
			return nil, nil
		}
		n, err := strconv.ParseInt(string(b), 10, int(bitSize))
		if err != nil {
			return nil, fmt.Errorf("%s value is not correct: %s", tpe, err)
		}
		return &n, nil
	}
	switch len(ranges) {
	case 2:
		min, err := parseInt(ranges[0].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("min %s", err)
		}
		max, err := parseInt(ranges[1].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("max %s", err)
		}
		if min != nil && max != nil && *max < *min {
			return nil, nil, fmt.Errorf("max %s value must be equal or large than min expect %d, current %d", tpe, min, max)
		}

		return min, max, nil
	case 1:
		min, err := parseInt(ranges[0].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("min %s", err)
		}
		return min, min, nil
	}
	return nil, nil, nil
}

func (validator *IntValidator) String() string {
	rule := rules.NewRule(validator.Names()[0])

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(strconv.Itoa(int(validator.BitSize)))),
	}

	if validator.Minimum != nil || validator.Maximum != nil {
		rule.Range = make([]*rules.RuleLit, 2)

		if validator.Minimum != nil {
			rule.Range[0] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%d", *validator.Minimum)),
			)
		}

		if validator.Maximum != nil {
			rule.Range[1] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%d", *validator.Maximum)),
			)
		}

		rule.ExclusiveLeft = validator.ExclusiveMinimum
		rule.ExclusiveRight = validator.ExclusiveMaximum
	}

	rule.ExclusiveLeft = validator.ExclusiveMinimum
	rule.ExclusiveRight = validator.ExclusiveMaximum

	if validator.MultipleOf != 0 {
		rule.Values = []*rules.RuleLit{
			rules.NewRuleLit([]byte("%" + fmt.Sprintf("%d", validator.MultipleOf))),
		}
	} else if validator.Enums != nil {
		for _, str := range validator.Enums {
			rule.Values = append(rule.Values, rules.NewRuleLit([]byte(str)))
		}
	}

	return string(rule.Bytes())
}

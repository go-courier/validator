package validator

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"unicode"
)

// @uint<BIT_SIZE>[from,to]
// @uint<BIT_SIZE>[from,to]
// @uint<BIT_SIZE>[from,to)
// @uint<BIT_SIZE>{%multipleOf}
//
// aliases:
// uint8 = uint<8>
// uint16 = uint<16>
// uint32 = uint<8>
// uint64 = uint<64>
// uint<53> for JavaScript https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER
type UintValidator struct {
	BitSize uint

	Minimum          uint64
	Maximum          uint64
	MultipleOf       uint64
	ExclusiveMaximum bool
	ExclusiveMinimum bool

	Enums map[uint64]string
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

func (validator *UintValidator) Validate(v interface{}) error {
	val := uint64(0)
	switch i := v.(type) {
	case uint8:
		val = uint64(i)
	case uint16:
		val = uint64(i)
	case uint:
		val = uint64(i)
	case uint32:
		val = uint64(i)
	case uint64:
		val = i
	default:
		return NewUnsupportedTypeError("uint64", reflect.TypeOf(v))
	}

	if validator.Enums != nil {
		if _, ok := validator.Enums[val]; !ok {
			return fmt.Errorf("unknown enumeration value %d", val)
		}
		return nil
	}

	if ((validator.ExclusiveMinimum && val == validator.Minimum) || val < validator.Minimum) ||
		((validator.ExclusiveMaximum && val == validator.Maximum) || val > validator.Maximum) {
		return fmt.Errorf("uint out of range %s，current：%d", validator, val)
	}

	if validator.MultipleOf != 0 {
		if val%validator.MultipleOf != 0 {
			return fmt.Errorf("uint value should be multiple of %d，current：%d", validator.MultipleOf, val)
		}
	}

	return nil
}

func (UintValidator) New(rule *Rule) (Validator, error) {
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
			return nil, NewSyntaxErrorf("unit parameter should be valid bit size, but got `%s`", bitSizeStr)
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

	if rule.Values != nil {
		if len(rule.Values) == 1 {
			mayBeMultipleOf := rule.Values[0].Bytes()
			if mayBeMultipleOf[0] == '%' {
				v := mayBeMultipleOf[1:]
				multipleOf, err := strconv.ParseUint(string(v), 10, int(validator.BitSize))
				if err != nil {
					return nil, NewSyntaxErrorf("multipleOf should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.MultipleOf = multipleOf
			}
		}

		if validator.MultipleOf == 0 {
			validator.Enums = map[uint64]string{}
			for _, v := range rule.Values {
				str := string(v.Bytes())
				enumValue, err := strconv.ParseUint(str, 10, int(validator.BitSize))
				if err != nil {
					return nil, NewSyntaxErrorf("enum should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.Enums[enumValue] = str
			}
		}
	}

	return validator, nil
}

func (validator *UintValidator) String() string {
	rule := NewRule(validator.Names()[0])

	rule.Params = []RuleNode{
		NewRuleLit([]byte(strconv.Itoa(int(validator.BitSize)))),
	}

	rule.Range = []*RuleLit{
		NewRuleLit([]byte(fmt.Sprintf("%d", validator.Minimum))),
		NewRuleLit([]byte(fmt.Sprintf("%d", validator.Maximum))),
	}

	rule.ExclusiveLeft = validator.ExclusiveMinimum
	rule.ExclusiveRight = validator.ExclusiveMaximum

	if validator.MultipleOf != 0 {
		rule.Values = []*RuleLit{
			NewRuleLit([]byte("%" + fmt.Sprintf("%d", validator.MultipleOf))),
		}
	} else if validator.Enums != nil {
		for _, str := range validator.Enums {
			rule.Values = append(rule.Values, NewRuleLit([]byte(str)))
		}
	}

	return string(rule.Bytes())
}

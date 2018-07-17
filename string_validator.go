package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-courier/validator/errors"
	"github.com/go-courier/validator/rules"
)

var (
	TargetStringLength = "string length"
)

type StrLenMode int

const (
	STR_LEN_MODE__LENGTH StrLenMode = iota
	STR_LEN_MODE__RUNE_COUNT
)

var strLenModes = map[StrLenMode]func(s string) uint64{
	STR_LEN_MODE__LENGTH: func(s string) uint64 {
		return uint64(len(s))
	},
	STR_LEN_MODE__RUNE_COUNT: func(s string) uint64 {
		return uint64(utf8.RuneCount([]byte(s)))
	},
}

func ParseStrLenMode(s string) (StrLenMode, error) {
	switch strings.ToLower(s) {
	case "rune_count":
		return STR_LEN_MODE__RUNE_COUNT, nil
	case "length", "":
		return STR_LEN_MODE__LENGTH, nil
	default:
		return STR_LEN_MODE__LENGTH, fmt.Errorf("unsupported string length mode")
	}
}

func (m StrLenMode) String() string {
	switch m {
	case STR_LEN_MODE__RUNE_COUNT:
		return "rune_count"
	default:
		return "length"
	}
}

// @string{VALUE_1,VALUE_2,VALUE_3}
// @string/regexp/
// @string<StrLenMode>[from,to]
// @string<StrLenMode>[length]
//
// aliases
// * char as string<rune_count>

/*
Validator for string

Rules:

ranges
	@string[min,max]
	@string[length]
	@string[1,10] // string length should large or equal than 1 and less or equal than 10
	@string[1,]  // string length should large or equal than 1 and less than the maxinum of int32
	@string[,1]  // string length should less than 1 and large or equal than 0
	@string[10]  // string length should be equal 10

enumeration
	@string{A,B,C} // should one of these values

regexp
	@string/\w+/ // string values should match \w+
since we use / as wrapper for regexp, we need to use \ to escape /

length mode in parameter
	@string<length> // use string length directly
	@string<rune_count> // use rune count as string length

composes
	@string<>[1,]

aliases:
	@char = @string<rune_count>
*/
type StringValidator struct {
	Enums   map[string]string
	Pattern *regexp.Regexp
	LenMode StrLenMode

	MinLength uint64
	MaxLength *uint64
}

func init() {
	ValidatorMgrDefault.Register(&StringValidator{})
}

func (StringValidator) Names() []string {
	return []string{"string", "char"}
}

var typString = reflect.TypeOf("")

func (validator *StringValidator) Validate(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if !rv.Type().ConvertibleTo(typString) {
		return errors.NewUnsupportedTypeError(rv.Type().String(), validator.String())
	}

	s := rv.Convert(typString).String()

	if validator.Enums != nil {
		if _, ok := validator.Enums[s]; !ok {
			values := make([]interface{}, 0)
			for _, v := range validator.Enums {
				values = append(values, v)
			}

			return &errors.NotInEnumError{
				Target:  TargetStringLength,
				Current: v,
				Enums:   values,
			}
		}
		return nil
	}

	if validator.Pattern != nil {
		if !validator.Pattern.MatchString(s) {
			return &errors.NotMatchError{
				Target:  TargetStringLength,
				Pattern: validator.Pattern,
				Current: v,
			}
		}
		return nil
	}

	strLen := strLenModes[validator.LenMode](s)

	if strLen < validator.MinLength {
		return &errors.OutOfRangeError{
			Target:  TargetStringLength,
			Current: v,
			Minimum: validator.MinLength,
		}
	}

	if validator.MaxLength != nil && strLen > *validator.MaxLength {
		return &errors.OutOfRangeError{
			Target:  TargetStringLength,
			Current: v,
			Maximum: validator.MaxLength,
		}
	}
	return nil
}

func (StringValidator) New(rule *Rule, mgr ValidatorMgr) (Validator, error) {
	validator := &StringValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, errors.NewSyntaxError("range mark of %s should not be `(` or `)`", validator.Names()[0])
	}

	if rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("string should only 1 parameter, but got %d", len(rule.Params))
		}
		lenMode, err := ParseStrLenMode(string(rule.Params[0].Bytes()))
		if err != nil {
			return nil, err
		}
		validator.LenMode = lenMode
	} else if rule.Name == "char" {
		validator.LenMode = STR_LEN_MODE__RUNE_COUNT
	}

	if rule.Pattern != nil {
		validator.Pattern = rule.Pattern
		return validator, validator.TypeCheck(rule)
	}

	if rule.Values != nil {
		validator.Enums = map[string]string{}
		for _, v := range rule.Values {
			str := string(v.Bytes())
			validator.Enums[str] = str
		}
	}

	if rule.Range != nil {
		min, max, err := UintRange(fmt.Sprintf("%s of string", validator.LenMode), 64, rule.Range...)
		if err != nil {
			return nil, err
		}
		validator.MinLength = min
		validator.MaxLength = max
	}

	return validator, validator.TypeCheck(rule)
}

func (validator *StringValidator) TypeCheck(rule *Rule) error {
	if rule.Type.Kind() == reflect.String {
		return nil
	}
	return errors.NewUnsupportedTypeError(rule.String(), validator.String())
}

func (validator *StringValidator) String() string {
	rule := rules.NewRule(validator.Names()[0])

	if validator.Enums != nil {
		for e := range validator.Enums {
			rule.Values = append(rule.Values, rules.NewRuleLit([]byte(e)))
		}
	}

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(validator.LenMode.String())),
	}

	if validator.Pattern != nil {
		rule.Pattern = validator.Pattern
	}

	rule.Range = RangeFromUint(validator.MinLength, validator.MaxLength)

	return string(rule.Bytes())
}

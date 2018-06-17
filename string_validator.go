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

func (StringValidator) Names() []string {
	return []string{"string", "char"}
}

func (validator *StringValidator) Validate(v interface{}) error {
	if rv, ok := v.(reflect.Value); ok && rv.CanInterface() {
		v = rv.Interface()
	}

	s, ok := v.(string)
	if !ok {
		return errors.NewUnsupportedTypeError(reflect.TypeOf(v), validator.String())
	}

	if validator.Enums != nil {
		if _, ok := validator.Enums[s]; !ok {
			return fmt.Errorf("unknown enumeration value %s", s)
		}
		return nil
	}

	if validator.Pattern != nil {
		if !validator.Pattern.MatchString(s) {
			return fmt.Errorf("string not match `%s`，current：%s", validator.Pattern, s)
		}
		return nil
	}

	strLen := strLenModes[validator.LenMode](s)

	if strLen < validator.MinLength || (validator.MaxLength != nil && strLen > *validator.MaxLength) {
		return fmt.Errorf("string length out of range %s，current：%d", validator, strLen)
	}
	return nil
}

func (StringValidator) New(rule *rules.Rule, tpe reflect.Type, mgr ValidatorMgr) (Validator, error) {
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
		return validator, validator.TypeCheck(tpe)
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

	return validator, validator.TypeCheck(tpe)
}

func (validator *StringValidator) TypeCheck(tpe reflect.Type) error {
	if tpe.Kind() == reflect.String {
		return nil
	}
	return errors.NewUnsupportedTypeError(tpe, validator.String())
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

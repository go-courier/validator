package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"
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
	s, ok := v.(string)
	if !ok {
		return NewUnsupportedTypeError("string", reflect.TypeOf(v))
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

func (StringValidator) New(rule *Rule) (Validator, error) {
	validator := &StringValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, NewSyntaxErrorf("range mark of %s should not be `(` or `)`", validator.Names()[0])
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
		return validator, nil
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

	return validator, nil
}

func (validator *StringValidator) String() string {
	rule := NewRule(validator.Names()[0])

	if validator.Enums != nil {
		for e := range validator.Enums {
			rule.Values = append(rule.Values, NewRuleLit([]byte(e)))
		}
	}

	rule.Params = []RuleNode{
		NewRuleLit([]byte(validator.LenMode.String())),
	}

	if validator.Pattern != nil {
		rule.Pattern = validator.Pattern
	}

	rule.Range = RangeFromUint(validator.MinLength, validator.MaxLength)

	return string(rule.Bytes())
}

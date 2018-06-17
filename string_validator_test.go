package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/go-courier/validator/rules"
)

func TestStringValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *StringValidator
	}{
		reflect.TypeOf(""): {
			{"@string[1,1000]", &StringValidator{
				MinLength: 1,
				MaxLength: ptr.Uint64(1000),
			}},
			{"@string[1,]", &StringValidator{
				MinLength: 1,
			}},
			{"@string<length>[1]", &StringValidator{
				MinLength: 1,
				MaxLength: ptr.Uint64(1),
			}},
			{"@char[1,]", &StringValidator{
				LenMode:   STR_LEN_MODE__RUNE_COUNT,
				MinLength: 1,
			}},
			{"@string<rune_count>[1,]", &StringValidator{
				LenMode:   STR_LEN_MODE__RUNE_COUNT,
				MinLength: 1,
			}},
			{"@string{KEY1,KEY2}", &StringValidator{
				Enums: map[string]string{
					"KEY1": "KEY1",
					"KEY2": "KEY2",
				},
			}},
			{`@string/^\w+/`, &StringValidator{
				Pattern: regexp.MustCompile(`^\w+`),
			}},
			{`@string/^\w+\/test/`, &StringValidator{
				Pattern: regexp.MustCompile(`^\w+/test`),
			}},
		},
	}

	for tpe, cases := range caseSet {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s %s|%s", tpe, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(rules.MustParseRuleString(c.rule), tpe, nil)
				assert.NoError(t, err)
				assert.Equal(t, c.expect, v)
			})
		}
	}
}

func TestStringValidator_NewFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(1): {
			"@string",
		},
		reflect.TypeOf(""): {
			"@string<length, 1>",
			"@string<unsupported>",
			"@string[1,0]",
			"@string[1,-2]",
			"@string[a,]",
			"@string[-1,1]",
			"@string(-1,1)",
		},
	}

	validator := &StringValidator{}

	for tpe := range invalidRules {
		for _, r := range invalidRules[tpe] {
			rule := rules.MustParseRuleString(r)

			t.Run(fmt.Sprintf("validate %s new failed: %s", tpe, rule.Bytes()), func(t *testing.T) {
				_, err := validator.New(rule, tpe, validatorFactory)
				assert.Error(t, err)
				t.Log(err)
			})
		}
	}
}

func TestStringValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *StringValidator
		desc      string
	}{
		{[]interface{}{reflect.ValueOf("a"), "aa", "aaa", "aaaa", "aaaaa"}, &StringValidator{
			MaxLength: ptr.Uint64(5),
		}, "less than"},
		{[]interface{}{"一", "一一", "一一一"}, &StringValidator{
			LenMode:   STR_LEN_MODE__RUNE_COUNT,
			MaxLength: ptr.Uint64(3),
		}, "char count less than"},
		{[]interface{}{"A", "B"}, &StringValidator{
			Enums: map[string]string{
				"A": "A",
				"B": "B",
			},
		}, "in enum"},
		{[]interface{}{"word", "word1"}, &StringValidator{
			Pattern: regexp.MustCompile(`^\w+`),
		}, "regexp matched"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				assert.NoError(t, c.validator.Validate(v))
			})
		}
	}
}

func TestStringValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *StringValidator
		desc      string
	}{
		{[]interface{}{"C", "D", "E"}, &StringValidator{
			Enums: map[string]string{
				"A": "A",
				"B": "B",
			},
		}, "enum not match"},
		{[]interface{}{"-word", "-word1"}, &StringValidator{
			Pattern: regexp.MustCompile(`^\w+`),
		}, "regexp not matched"},
		{[]interface{}{1, 1.1, reflect.ValueOf(1)}, &StringValidator{
			MinLength: 5,
		}, "unsupported types"},
		{[]interface{}{"a", "aa", "aaa", "aaaa"}, &StringValidator{
			MinLength: 5,
		}, "too small"},
		{[]interface{}{"aa", "aaa", "aaaa", "aaaaa"}, &StringValidator{
			MaxLength: ptr.Uint64(1),
		}, "too large"},
		{[]interface{}{"字符太多"}, &StringValidator{
			LenMode:   STR_LEN_MODE__RUNE_COUNT,
			MaxLength: ptr.Uint64(3),
		}, "too many chars"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				assert.Error(t, err)
				t.Log(err)
			})
		}
	}
}

package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/stretchr/testify/require"
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

	for typ, cases := range caseSet {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s %s|%s", typ, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(MustParseRuleStringWithType(c.rule, typesutil.FromRType(typ)), nil)
				require.NoError(t, err)
				require.Equal(t, c.expect, v)
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

	for typ := range invalidRules {
		for _, r := range invalidRules[typ] {
			rule := MustParseRuleStringWithType(r, typesutil.FromRType(typ))

			t.Run(fmt.Sprintf("validate %s new failed: %s", typ, rule.Bytes()), func(t *testing.T) {
				_, err := validator.New(rule, ValidatorMgrDefault)
				require.Error(t, err)
				t.Log(err)
			})
		}
	}
}

func TestStringValidator_Validate(t *testing.T) {
	type String string

	cases := []struct {
		values    []interface{}
		validator *StringValidator
		desc      string
	}{
		{[]interface{}{reflect.ValueOf("a"), String("aa"), "aaa", "aaaa", "aaaaa"}, &StringValidator{
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
				require.NoError(t, c.validator.Validate(v))
			})
		}
	}
}

func TestStringValidator_ValidateFailed(t *testing.T) {
	type String string

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
		{[]interface{}{1.1, reflect.ValueOf(1.1)}, &StringValidator{
			MinLength: 5,
		}, "unsupported types"},
		{[]interface{}{"a", "aa", String("aaa"), []byte("aaaa")}, &StringValidator{
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
				require.Error(t, err)
				t.Log(err)
			})
		}
	}
}

package validator

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

func TestStringValidator_New(t *testing.T) {
	cases := []struct {
		rule   string
		expect *StringValidator
	}{
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
	}

	for i := range cases {
		c := cases[i]

		t.Run(c.rule+"|"+c.expect.String(), func(t *testing.T) {
			rule := MustParseRuleString(c.rule)
			v, err := c.expect.New(rule)
			assert.NoError(t, err)
			assert.Equal(t, c.expect, v)
		})
	}
}

func TestStringValidator_NewFailed(t *testing.T) {
	invalidRules := []string{
		"@string<length, 1>",
		"@string<unsupported>",
		"@string[1,0]",
		"@string[1,-2]",
		"@string[a,]",
		"@string[-1,1]",
		"@string(-1,1)",
	}

	for i := range invalidRules {
		rule := MustParseRuleString(invalidRules[i])
		validator := &StringValidator{}

		t.Run(fmt.Sprintf("validate new failed: %s", rule.Bytes()), func(t *testing.T) {
			_, err := validator.New(rule)
			assert.Error(t, err)
		})
	}
}

func TestStringValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *StringValidator
		desc      string
	}{
		{[]interface{}{"a", "aa", "aaa", "aaaa", "aaaaa"}, &StringValidator{
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
		{[]interface{}{1, 1.1}, &StringValidator{
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

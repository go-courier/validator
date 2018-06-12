package validator

import (
	"fmt"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

func TestMapValidator_New(t *testing.T) {
	cases := []struct {
		rule   string
		expect *MapValidator
	}{
		{"@map[1,1000]", &MapValidator{
			MinProperties: 1,
			MaxProperties: ptr.Uint64(1000),
		}},
		{"@map<,@map[1,2]>[1,]", &MapValidator{
			MinProperties: 1,
			ElemRule:      MustParseRuleString("@map[1,2]"),
		}},
		{"@map<@email,@map[1,2]>[1,]", &MapValidator{
			MinProperties: 1,
			KeyRule:       MustParseRuleString("@email"),
			ElemRule:      MustParseRuleString("@map[1,2]"),
		}},
	}

	for i := range cases {
		c := cases[i]

		t.Run(c.rule+"|"+c.expect.String(), func(t *testing.T) {
			rule := MustParseRuleString(c.rule)
			c.expect.Rule = rule
			v, err := c.expect.New(rule)
			assert.NoError(t, err)
			assert.Equal(t, c.expect, v)
		})
	}
}

func TestMapValidator_NewFailed(t *testing.T) {
	invalidRules := []string{
		"@map<1,>",
		"@map<,2>",
		"@map<1,2,3>",
		"@map[1,0]",
		"@map[1,-2]",
		"@map[a,]",
		"@map[-1,1]",
		"@map(-1,1)",
	}

	for i := range invalidRules {
		rule := MustParseRuleString(invalidRules[i])
		validator := &MapValidator{}

		t.Run(fmt.Sprintf("validate new failed: %s", rule.Bytes()), func(t *testing.T) {
			_, err := validator.New(rule)
			assert.Error(t, err)
		})
	}
}

func TestMapValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *MapValidator
		desc      string
	}{
		{[]interface{}{
			map[string]string{"1": "", "2": ""},
			map[string]string{"1": "", "2": "", "3": ""},
			map[string]string{"1": "", "2": "", "3": "", "4": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
		}, "in range"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				assert.NoError(t, c.validator.Validate(v))
			})
		}
	}
}

func TestMapValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *MapValidator
		desc      string
	}{
		{[]interface{}{
			1,
			"s",
			true,
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
		}, "unsupported type"},
		{[]interface{}{
			map[string]string{"1": ""},
			map[string]string{"1": "", "2": "", "3": "", "4": "", "5": ""},
			map[string]string{"1": "", "2": "", "3": "", "4": "", "5": "", "6": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
		}, "out of range"},
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

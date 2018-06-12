package validator

import (
	"fmt"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

func TestSliceValidator_New(t *testing.T) {
	cases := []struct {
		rule   string
		expect *SliceValidator
	}{
		{"@slice[1,1000]", &SliceValidator{
			MinItems: 1,
			MaxItems: ptr.Uint64(1000),
		}},
		{"@slice<@string[1,2]>[1,]", &SliceValidator{
			MinItems: 1,
			ElemRule: MustParseRuleString("@string[1,2]"),
		}},
		{"@slice[1]", &SliceValidator{
			MinItems: 1,
			MaxItems: ptr.Uint64(1),
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

func TestSliceValidator_NewFailed(t *testing.T) {
	invalidRules := []string{
		"@slice<1>",
		"@slice<1,2,4>",
		"@slice[1,0]",
		"@slice[1,-2]",
		"@slice[a,]",
		"@slice[-1,1]",
		"@slice(1,1)",
	}

	for i := range invalidRules {
		rule := MustParseRuleString(invalidRules[i])
		validator := &SliceValidator{}

		t.Run(fmt.Sprintf("validate new failed: %s", rule.Bytes()), func(t *testing.T) {
			_, err := validator.New(rule)
			assert.Error(t, err)
		})
	}
}

func TestSliceValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *SliceValidator
		desc      string
	}{
		{[]interface{}{
			[]string{"1", "2"},
			[]string{"1", "2", "3"},
			[]string{"1", "2", "3", "4"},
		}, &SliceValidator{
			MinItems: 2,
			MaxItems: ptr.Uint64(4),
		}, "in range"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				assert.NoError(t, err)
			})
		}
	}
}

func TestSliceValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *SliceValidator
		desc      string
	}{
		{[]interface{}{
			[]string{"1"},
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5", "6"},
		}, &SliceValidator{
			MinItems: 2,
			MaxItems: ptr.Uint64(4),
		}, "out of range"},
		{[]interface{}{
			1,
		}, &SliceValidator{
			MinItems: 2,
			MaxItems: ptr.Uint64(4),
		}, "unsupported type"},
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

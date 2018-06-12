package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUintValidator_New(t *testing.T) {
	cases := []struct {
		rule   string
		expect *UintValidator
	}{
		{"@uint[1,1000]", &UintValidator{
			Minimum: 1,
			Maximum: 1000,
		}},
		{"@uint[1,1000)", &UintValidator{
			Minimum:          1,
			Maximum:          1000,
			ExclusiveMaximum: true,
		}},
		{"@uint(1,1000]", &UintValidator{
			Minimum:          1,
			Maximum:          1000,
			ExclusiveMinimum: true,
		}},
		{"@uint[1,]", &UintValidator{
			Minimum: 1,
			Maximum: MaxUint(32),
		}},
		{"@uint16{1,2}", &UintValidator{
			BitSize: 16,
			Enums: map[uint64]string{
				1: "1",
				2: "2",
			},
		}},
		{"@uint<53>", &UintValidator{
			BitSize: 53,
			Maximum: MaxUint(53),
		}},
		{"@uint16{%2}", &UintValidator{
			BitSize:    16,
			MultipleOf: 2,
		}},
	}

	for _, c := range cases {
		c.expect.SetDefaults()

		t.Run(fmt.Sprintf("%s|%s", c.rule, c.expect.String()), func(t *testing.T) {
			rule := MustParseRuleString(c.rule)
			v, err := c.expect.New(rule)
			assert.NoError(t, err)
			assert.Equal(t, c.expect, v)
		})
	}
}

func TestUintValidator_ParseFailed(t *testing.T) {
	invalidRules := []string{
		"@uint<32,2123>",
		"@uint<@string>",
		"@uint<66>",
		"@uint[1,0]",
		"@uint[1,-2]",
		"@uint[a,]",
		"@uint[-1,1]",
		"@uint(-1,1)",
		`@uint8{%a}`,
		`@uint8{A,B,C}`,
	}

	for i := range invalidRules {
		rule := MustParseRuleString(invalidRules[i])
		validator := &UintValidator{}

		t.Run(fmt.Sprintf("validate new failed: %s", rule.Bytes()), func(t *testing.T) {
			_, err := validator.New(rule)
			assert.Error(t, err)
		})
	}
}

func TestUintValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *UintValidator
		desc      string
	}{
		{[]interface{}{uint(1), uint(2), uint(3)}, &UintValidator{
			Enums: map[uint64]string{
				1: "1",
				2: "2",
				3: "3",
			},
		}, "in enum"},
		{[]interface{}{uint(2), uint(3), uint(4)}, &UintValidator{
			Minimum: 2,
			Maximum: 4,
		}, "in range"},
		{[]interface{}{uint8(2), uint16(3), uint32(4), uint64(4)}, &UintValidator{
			Minimum: 2,
			Maximum: 4,
		}, "uint types"},
		{[]interface{}{uint64(2), uint64(3), uint64(4)}, &UintValidator{
			BitSize: 64,
			Minimum: 2,
			Maximum: 4,
		}, "in range"},
		{[]interface{}{uint(2), uint(4), uint(6)}, &UintValidator{
			MultipleOf: 2,
		}, "multiple of"},
	}

	for _, c := range cases {
		c.validator.SetDefaults()
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				assert.NoError(t, c.validator.Validate(v))
			})
		}
	}
}

func TestUintValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *UintValidator
		desc      string
	}{
		{[]interface{}{2, "string"}, &UintValidator{
			BitSize: 64,
		}, "unsupported type"},
		{[]interface{}{uint(4), uint(5), uint(6)}, &UintValidator{
			Enums: map[uint64]string{
				1: "1",
				2: "2",
				3: "3",
			},
		}, "not in enum"},
		{[]interface{}{uint(1), uint(4), uint(5)}, &UintValidator{
			Minimum:          2,
			Maximum:          4,
			ExclusiveMaximum: true,
		}, "not in range"},
		{[]interface{}{uint(1), uint(3), uint(5)}, &UintValidator{
			MultipleOf: 2,
		}, "not multiple of"},
	}

	for _, c := range cases {
		c.validator.SetDefaults()
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				assert.Error(t, err)
				t.Log(err)
			})
		}
	}
}

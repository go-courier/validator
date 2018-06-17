package validator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/go-courier/validator/rules"
)

func TestIntValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *IntValidator
	}{
		reflect.TypeOf(int8(1)): {
			{"@int8[1,]", &IntValidator{
				BitSize: 8,
				Minimum: ptr.Int64(1),
				Maximum: ptr.Int64(MaxInt(8)),
			}},
		},
		reflect.TypeOf(int16(1)): {
			{"@int16[1,]", &IntValidator{
				BitSize: 16,
				Minimum: ptr.Int64(1),
				Maximum: ptr.Int64(MaxInt(16)),
			}},
		},
		reflect.TypeOf(int32(1)): {
			{"@int[1,]", &IntValidator{
				Minimum: ptr.Int64(1),
				Maximum: ptr.Int64(MaxInt(32)),
			}},
			{"@int[1,1000)", &IntValidator{
				Minimum:          ptr.Int64(1),
				Maximum:          ptr.Int64(1000),
				ExclusiveMaximum: true,
			}},
			{"@int(1,1000]", &IntValidator{
				Minimum:          ptr.Int64(1),
				Maximum:          ptr.Int64(1000),
				ExclusiveMinimum: true,
			}},
			{"@int[1,]", &IntValidator{
				Minimum: ptr.Int64(1),
				Maximum: ptr.Int64(MaxInt(32)),
			}},
			{"@int[1]", &IntValidator{
				Minimum: ptr.Int64(1),
				Maximum: ptr.Int64(1),
			}},
			{"@int[,1]", &IntValidator{
				Maximum: ptr.Int64(1),
			}},
			{"@int16{1,2}", &IntValidator{
				BitSize: 16,
				Enums: map[int64]string{
					1: "1",
					2: "2",
				},
			}},
			{"@int16{%2}", &IntValidator{
				BitSize:    16,
				MultipleOf: 2,
			}},
		},
		reflect.TypeOf(int64(1)): {
			{"@int64[1,1000]", &IntValidator{
				BitSize: 64,
				Minimum: ptr.Int64(1),
				Maximum: ptr.Int64(1000),
			}},
			{"@int<53>", &IntValidator{
				BitSize: 53,
				Maximum: ptr.Int64(MaxInt(53)),
			}},
		},
	}

	for tpe, cases := range caseSet {
		for _, c := range cases {
			c.expect.SetDefaults()

			t.Run(fmt.Sprintf("%s %s|%s", tpe, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(rules.MustParseRuleString(c.rule), tpe, nil)
				assert.NoError(t, err)
				assert.Equal(t, c.expect, v)
			})
		}
	}
}

func TestIntValidator_NewFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(float32(1)): {
			"@int16",
		},
		reflect.TypeOf(int8(1)): {
			"@int16",
		},
		reflect.TypeOf(int16(1)): {
			"@int",
		},
		reflect.TypeOf(int32(1)): {
			"@int64",
		},
		reflect.TypeOf(int(1)): {
			"@int<32,2123>",
			"@int<@string>",
			"@int<66>",
			"@int[1,0]",
			"@int[1,-2]",
			"@int[a,]",
			"@int[,a]",
			"@int[a]",
			`@int8{%a}`,
			`@int8{A,B,C}`,
		},
	}

	validator := &IntValidator{}

	for tpe := range invalidRules {
		for _, r := range invalidRules[tpe] {
			rule := rules.MustParseRuleString(r)

			t.Run(fmt.Sprintf("validate %s new failed: %s", tpe, rule.Bytes()), func(t *testing.T) {
				_, err := validator.New(rule, tpe, nil)
				assert.Error(t, err)
			})
		}
	}
}

func TestIntValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *IntValidator
		desc      string
	}{
		{[]interface{}{reflect.ValueOf(int(1)), int(2), int(3)}, &IntValidator{
			Enums: map[int64]string{
				1: "1",
				2: "2",
				3: "3",
			},
		}, "in enum"},
		{[]interface{}{int(2), int(3), int(4)}, &IntValidator{
			Minimum: ptr.Int64(2),
			Maximum: ptr.Int64(4),
		}, "in range"},
		{[]interface{}{int8(2), int16(3), int32(4), int64(4)}, &IntValidator{
			Minimum: ptr.Int64(2),
			Maximum: ptr.Int64(4),
		}, "int types"},
		{[]interface{}{int64(2), int64(3), int64(4)}, &IntValidator{
			BitSize: 64,
			Minimum: ptr.Int64(2),
			Maximum: ptr.Int64(4),
		}, "in range"},
		{[]interface{}{int(2), int(4), int(6)}, &IntValidator{
			MultipleOf: 2,
		}, "multiple of"},
	}
	for i := range cases {
		c := cases[i]
		c.validator.SetDefaults()
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				assert.NoError(t, c.validator.Validate(v))
			})
		}
	}
}

func TestIntValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *IntValidator
		desc      string
	}{
		{[]interface{}{uint(2), "string", reflect.ValueOf("1")}, &IntValidator{
			BitSize: 64,
		}, "unsupported type"},
		{[]interface{}{int(4), int(5), int(6)}, &IntValidator{
			Enums: map[int64]string{
				1: "1",
				2: "2",
				3: "3",
			},
		}, "not in enum"},
		{[]interface{}{int(1), int(4), int(5)}, &IntValidator{
			Minimum:          ptr.Int64(2),
			Maximum:          ptr.Int64(4),
			ExclusiveMaximum: true,
		}, "not in range"},
		{[]interface{}{int(1), int(3), int(5)}, &IntValidator{
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

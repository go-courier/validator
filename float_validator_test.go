package validator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/require"

	"github.com/go-courier/validator/types"
)

func TestFloatValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *FloatValidator
	}{
		reflect.TypeOf(float32(1.1)): {
			{"@float[1,1000]", &FloatValidator{
				Minimum: ptr.Float64(1),
				Maximum: ptr.Float64(1000),
			}},
		},
		reflect.TypeOf(float64(1.1)): {
			{"@float[1,1000]", &FloatValidator{
				Minimum: ptr.Float64(1),
				Maximum: ptr.Float64(1000),
			}},
			{"@float32[1,1000]", &FloatValidator{
				Minimum: ptr.Float64(1),
				Maximum: ptr.Float64(1000),
			}},
			{"@double[1,1000]", &FloatValidator{
				MaxDigits: 15,
				Minimum:   ptr.Float64(1),
				Maximum:   ptr.Float64(1000),
			}},
			{"@float64[1,1000]", &FloatValidator{
				MaxDigits: 15,
				Minimum:   ptr.Float64(1),
				Maximum:   ptr.Float64(1000),
			}},
			{"@float(1,1000]", &FloatValidator{
				Minimum:          ptr.Float64(1),
				ExclusiveMinimum: true,
				Maximum:          ptr.Float64(1000),
			}},
			{"@float[.1,]", &FloatValidator{
				Minimum: ptr.Float64(.1),
			}},
			{"@float[,-1]", &FloatValidator{
				Maximum: ptr.Float64(-1),
			}},
			{"@float[-1]", &FloatValidator{
				Minimum: ptr.Float64(-1),
				Maximum: ptr.Float64(-1),
			}},
			{"@float{1,2}", &FloatValidator{
				Enums: map[float64]string{
					1: "1",
					2: "2",
				},
			}},
			{"@float{%2.2}", &FloatValidator{
				MultipleOf: 2.2,
			}},
			{"@float<10,3>[1.333,2.333]", &FloatValidator{
				MaxDigits:     10,
				DecimalDigits: ptr.Uint(3),
				Minimum:       ptr.Float64(1.333),
				Maximum:       ptr.Float64(2.333),
			}},
		},
	}

	for typ, cases := range caseSet {
		for _, c := range cases {
			c.expect.SetDefaults()

			t.Run(fmt.Sprintf("%s %s|%s", typ, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(MustParseRuleStringWithType(c.rule, types.FromRType(typ)), nil)
				require.NoError(t, err)
				require.Equal(t, c.expect, v)
			})
		}
	}
}

func TestFloatValidator_NewFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(int(1)): {
			`@float64`,
		},
		reflect.TypeOf(float32(1.1)): {
			`@float64`,
			`@double`,
			`@float<9>`,
		},
		reflect.TypeOf(float64(1.1)): {
			"@float<11,22,33>",
			"@float<32,2123>",
			"@float<@string>",
			"@float<66>",
			"@float<7,7>",
			"@float[1,0]",
			"@float[1,-2]",
			"@float<7,2>[1.333,2]",
			"@float<7,2>[111111.33,]",
			"@float[a,]",
			"@float[,a]",
			"@float[a]",
			`@float{%a}`,
			`@float{A,B,C}`,
		},
	}

	validator := &FloatValidator{}

	for typ := range invalidRules {
		for _, r := range invalidRules[typ] {
			rule := MustParseRuleStringWithType(r, types.FromRType(typ))

			t.Run(fmt.Sprintf("validate %s new failed: %s", typ, rule.Bytes()), func(t *testing.T) {
				_, err := validator.New(rule, nil)
				require.Error(t, err)
				t.Log(err)
			})
		}
	}
}

func TestFloatValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *FloatValidator
		desc      string
	}{
		{[]interface{}{reflect.ValueOf(float64(1)), float64(2), float64(3)}, &FloatValidator{
			Enums: map[float64]string{
				1: "1",
				2: "2",
				3: "3",
			},
		}, "in enum"},
		{[]interface{}{float64(2), float64(3), float64(4)}, &FloatValidator{
			Minimum: ptr.Float64(2),
			Maximum: ptr.Float64(4),
		}, "in range"},
		{[]interface{}{float64(2), float64(3), float64(4), float64(4)}, &FloatValidator{
			Minimum: ptr.Float64(2),
			Maximum: ptr.Float64(4),
		}, "int types"},
		{[]interface{}{float32(2), float32(3), float32(4)}, &FloatValidator{
			Minimum: ptr.Float64(2),
			Maximum: ptr.Float64(4),
		}, "in range"},
		{[]interface{}{-2.2, 4.4, -6.6}, &FloatValidator{
			MultipleOf: 2.2,
		}, "multiple of"},
	}
	for i := range cases {
		c := cases[i]
		c.validator.SetDefaults()
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				require.NoError(t, c.validator.Validate(v))
			})
		}
	}
}

func TestFloatValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *FloatValidator
		desc      string
	}{
		{[]interface{}{
			uint(2),
			"string",
			reflect.ValueOf("1"),
		}, &FloatValidator{}, "unsupported type"},
		{[]interface{}{1.11, 1.22, float64(111111), float64(222221), 222.33}, &FloatValidator{
			MaxDigits:     5,
			DecimalDigits: ptr.Uint(1),
		}, "digits out out range range"},
		{[]interface{}{float64(4), float64(5), float64(6)}, &FloatValidator{
			Enums: map[float64]string{
				1: "1",
				2: "2",
				3: "3",
			},
		}, "not in enum"},
		{[]interface{}{float64(1), float64(4), float64(5)}, &FloatValidator{
			Minimum:          ptr.Float64(2),
			Maximum:          ptr.Float64(4),
			ExclusiveMaximum: true,
		}, "not in range"},
		{[]interface{}{1.1, 1.2, 1.3}, &FloatValidator{
			MultipleOf: 2,
		}, "not multiple of"},
	}

	for _, c := range cases {
		c.validator.SetDefaults()
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				require.Error(t, err)
				t.Log(err)
			})
		}
	}
}

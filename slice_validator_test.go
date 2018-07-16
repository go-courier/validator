package validator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/require"

	"github.com/go-courier/validator/rules"
)

func TestSliceValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *SliceValidator
	}{
		reflect.TypeOf([]string{}): {
			{"@slice[1,1000]", &SliceValidator{
				MinItems: 1,
				MaxItems: ptr.Uint64(1000),
			}},
			{"@slice<@string[1,2]>[1,]", &SliceValidator{
				MinItems:      1,
				ElemValidator: ValidatorMgrDefault.MustCompile([]byte("@string[1,2]"), reflect.TypeOf(""), nil),
			}},
			{"@slice[1]", &SliceValidator{
				MinItems: 1,
				MaxItems: ptr.Uint64(1),
			}},
		},
	}

	for tpe, cases := range caseSet {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s %s|%s", tpe, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(rules.MustParseRuleString(c.rule), tpe, ValidatorMgrDefault)
				require.NoError(t, err)
				require.Equal(t, c.expect, v)
			})
		}
	}
}

func TestSliceValidator_NewFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(""): {
			"@slice[2]",
		},
		reflect.TypeOf([1]string{}): {
			"@slice[2]",
		},
		reflect.TypeOf([]string{}): {
			"@slice<1>",
			"@slice<1,2,4>",
			"@slice[1,0]",
			"@slice[1,-2]",
			"@slice[a,]",
			"@slice[-1,1]",
			"@slice(1,1)",
			"@slice<@unknown>",
			"@array[1,2]",
		},
	}

	validator := &SliceValidator{}

	for tpe := range invalidRules {
		for _, r := range invalidRules[tpe] {
			rule := rules.MustParseRuleString(r)

			t.Run(fmt.Sprintf("validate %s new failed: %s", tpe, rule.Bytes()), func(t *testing.T) {
				_, err := validator.New(rule, tpe, ValidatorMgrDefault)
				require.Error(t, err)
				t.Log(err)
			})
		}
	}
}

func TestSliceValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *SliceValidator
		desc      string
	}{
		{[]interface{}{
			reflect.ValueOf([]string{"1", "2"}),
			[]string{"1", "2", "3"},
			[]string{"1", "2", "3", "4"},
		}, &SliceValidator{
			MinItems: 2,
			MaxItems: ptr.Uint64(4),
		}, "in range"},
		{[]interface{}{
			[]string{"1", "2"},
			[]string{"1", "2", "3"},
			[]string{"1", "2", "3", "4"},
		}, &SliceValidator{
			MinItems:      2,
			MaxItems:      ptr.Uint64(4),
			ElemValidator: ValidatorMgrDefault.MustCompile([]byte("@string[0,]"), reflect.TypeOf(""), nil),
		}, "elem validate"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				require.NoError(t, err)
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
			reflect.ValueOf(1),
			1,
		}, &SliceValidator{
			MinItems: 2,
			MaxItems: ptr.Uint64(4),
		}, "unsupported type"},
		{[]interface{}{
			[]string{"1", "2"},
			[]string{"1", "2", "3"},
			[]string{"1", "2", "3", "4"},
		}, &SliceValidator{
			MinItems:      2,
			MaxItems:      ptr.Uint64(4),
			ElemValidator: ValidatorMgrDefault.MustCompile([]byte("@string[2,]"), reflect.TypeOf(""), nil),
		}, "elem validate failed"},
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

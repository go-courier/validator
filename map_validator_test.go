package validator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/go-courier/validator/rules"
)

func TestMapValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *MapValidator
	}{
		reflect.TypeOf(map[string]string{}): {
			{"@map[1,1000]", &MapValidator{
				MinProperties: 1,
				MaxProperties: ptr.Uint64(1000),
			}},
		},
		reflect.TypeOf(map[string]map[string]string{}): {
			{"@map<,@map[1,2]>[1,]", &MapValidator{
				MinProperties: 1,
				ElemValidator: validatorFactory.MustCompile([]byte("@map[1,2]"), reflect.TypeOf(map[string]string{}), nil),
			}},
			{"@map<@string[0,],@map[1,2]>[1,]", &MapValidator{
				MinProperties: 1,
				KeyValidator:  validatorFactory.MustCompile([]byte("@string[0,]"), reflect.TypeOf(""), nil),
				ElemValidator: validatorFactory.MustCompile([]byte("@map[1,2]"), reflect.TypeOf(map[string]string{}), nil),
			}},
		},
	}

	for tpe, cases := range caseSet {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s %s|%s", tpe, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(rules.MustParseRuleString(c.rule), tpe, validatorFactory)
				assert.NoError(t, err)
				assert.Equal(t, c.expect, v)
			})
		}
	}
}

func TestMapValidator_NewFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf([]string{}): {
			"@map",
		},
		reflect.TypeOf(map[string]string{}): {
			"@map<1,>",
			"@map<,2>",
			"@map<1,2,3>",
			"@map[1,0]",
			"@map[1,-2]",
			"@map[a,]",
			"@map[-1,1]",
			"@map(-1,1)",
			"@map<@unknown,>",
			"@map<,@unknown>",
			"@map<@string[0,],@unknown>",
		},
	}

	validator := &MapValidator{}

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
		{[]interface{}{
			reflect.ValueOf(map[string]string{"1": "", "2": ""}),
			map[string]string{"1": "", "2": "", "3": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
			KeyValidator:  validatorFactory.MustCompile([]byte("@string[1,]"), reflect.TypeOf("1"), nil),
			ElemValidator: validatorFactory.MustCompile([]byte("@string[1,]?"), reflect.TypeOf("1"), nil),
		}, "key value validate"},
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
			reflect.ValueOf(1),
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
		{[]interface{}{
			map[string]string{"1": "", "2": ""},
			map[string]string{"1": "", "2": "", "3": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
			KeyValidator:  validatorFactory.MustCompile([]byte("@string[2,]"), reflect.TypeOf(""), nil),
			ElemValidator: validatorFactory.MustCompile([]byte("@string[2,]"), reflect.TypeOf(""), nil),
		}, "key elem validate failed"},
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

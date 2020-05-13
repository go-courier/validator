package validator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/stretchr/testify/require"
)

func TestSliceValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *SliceValidator
	}{
		reflect.TypeOf([]string{}): {
			{"@slice[1,1000]", &SliceValidator{
				MinItems:      1,
				MaxItems:      ptr.Uint64(1000),
				ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte(""), typesutil.FromRType(reflect.TypeOf("")), nil),
			}},
			{"@slice<@string[1,2]>[1,]", &SliceValidator{
				MinItems:      1,
				ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[1,2]"), typesutil.FromRType(reflect.TypeOf("")), nil),
			}},
			{"@slice[1]", &SliceValidator{
				MinItems:      1,
				MaxItems:      ptr.Uint64(1),
				ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte(""), typesutil.FromRType(reflect.TypeOf("")), nil),
			}},
		},
	}

	for typ, cases := range caseSet {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s %s|%s", typ, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), MustParseRuleStringWithType(c.rule, typesutil.FromRType(typ)))
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

	for typ := range invalidRules {
		for _, r := range invalidRules[typ] {
			rule := MustParseRuleStringWithType(r, typesutil.FromRType(typ))

			t.Run(fmt.Sprintf("validate %s new failed: %s", typ, rule.Bytes()), func(t *testing.T) {
				_, err := validator.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), rule)
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
			ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[0,]"), typesutil.FromRType(reflect.TypeOf("")), nil),
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
			[]string{"1", "2"},
			[]string{"1", "2", "3"},
			[]string{"1", "2", "3", "4"},
		}, &SliceValidator{
			MinItems:      2,
			MaxItems:      ptr.Uint64(4),
			ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[2,]"), typesutil.FromRType(reflect.TypeOf("")), nil),
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

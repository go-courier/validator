package validator

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-courier/ptr"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorLoader(t *testing.T) {
	type SomeStruct struct {
		PtrString *string
		String    string
	}

	var val *string
	someStruct := &SomeStruct{}

	cases := []struct {
		valuesPass   []interface{}
		valuesFailed []interface{}
		rule         string
		typ          reflect.Type
		validator    *ValidatorLoader
	}{
		{
			[]interface{}{
				reflect.ValueOf(someStruct).Elem().FieldByName("String"),
				"1",
			},
			[]interface{}{"222"},
			"@string[1,2] = '1'",
			reflect.TypeOf(""),
			&ValidatorLoader{
				Optional:        true,
				DefaultValue:    []byte("1"),
				PreprocessStage: PreprocessSkip,
			},
		},
		{
			[]interface{}{
				Duration(1 * time.Second),
				Duration(1 * time.Second),
			},
			[]interface{}{},
			"@string",
			reflect.TypeOf(Duration(1 * time.Second)),
			&ValidatorLoader{
				PreprocessStage: PreprocessString,
			},
		},
		{
			[]interface{}{
				val,
				reflect.ValueOf(someStruct).Elem().FieldByName("Value"),
				reflect.ValueOf(val),
				ptr.String("1"),
			},
			[]interface{}{
				ptr.String("222"),
			},
			"@string[1,2] = 2",
			reflect.TypeOf(ptr.String("")),
			&ValidatorLoader{
				Optional:        true,
				DefaultValue:    []byte("2"),
				PreprocessStage: PreprocessPtr,
			},
		},
		{
			[]interface{}{
				ptr.String("1"),
				ptr.String("22"),
			},
			[]interface{}{
				ptr.String(""),
				(*string)(nil),
			},
			"@string[1,2]",
			reflect.TypeOf(ptr.String("")),
			&ValidatorLoader{
				PreprocessStage: PreprocessPtr,
			},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s %s", c.typ, c.rule), func(t *testing.T) {
			validator, err := ValidatorMgrDefault.Compile(context.Background(), []byte(c.rule), typesutil.FromRType(c.typ), nil)
			require.NoError(t, err)
			if err != nil {
				return
			}

			loader := validator.(*ValidatorLoader)

			require.Equal(t, c.validator.Optional, loader.Optional)
			require.Equal(t, c.validator.PreprocessStage, loader.PreprocessStage)

			require.Equal(t, c.validator.DefaultValue, loader.DefaultValue)

			for _, v := range c.valuesPass {
				err := loader.Validate(v)
				require.NoError(t, err)
			}

			for _, v := range c.valuesFailed {
				err := loader.Validate(v)
				require.Error(t, err)
			}
		})
	}
}

func TestNewValidatorLoaderFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(1): {
			"@string",
			"@int[1,2] = s",
		},
		reflect.TypeOf(""): {
			"@string<length, 1>",
			"@string[1,2] = 123",
		},
		reflect.TypeOf(Duration(1)): {
			"@string[,10] = 2ss",
		},
	}

	for typ := range invalidRules {
		for _, r := range invalidRules[typ] {
			t.Run(fmt.Sprintf("%s validate %s", typ, r), func(t *testing.T) {
				_, err := ValidatorMgrDefault.Compile(context.Background(), []byte(r), typesutil.FromRType(typ), nil)
				require.Error(t, err)
				t.Log(err)
			})
		}
	}
}

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(data []byte) error {
	dur, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

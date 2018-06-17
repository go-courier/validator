package validator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

var validatorFactory = NewValidatorFactory(BuiltInValidators...)

func TestNewValidatorLoader(t *testing.T) {
	type SomeStruct struct {
		Value *string
	}

	var val *string
	someStruct := &SomeStruct{}

	cases := []struct {
		valuesPass   []interface{}
		valuesFailed []interface{}
		rule         string
		tpe          reflect.Type
		validator    *ValidatorLoader
	}{
		{
			[]interface{}{"", "1"},
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
		t.Run(fmt.Sprintf("%s %s", c.tpe, c.rule), func(t *testing.T) {
			validator, err := validatorFactory.Compile([]byte(c.rule), c.tpe, nil)
			assert.NoError(t, err)
			if err != nil {
				return
			}

			loader := validator.(*ValidatorLoader)

			assert.Equal(t, c.validator.Optional, loader.Optional)
			assert.Equal(t, c.validator.PreprocessStage, loader.PreprocessStage)

			assert.Equal(t, c.validator.DefaultValue, loader.DefaultValue)

			for _, v := range c.valuesPass {
				err := loader.Validate(v)
				assert.NoError(t, err)
			}

			for _, v := range c.valuesFailed {
				err := loader.Validate(v)
				assert.Error(t, err)
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

	for tpe := range invalidRules {
		for _, r := range invalidRules[tpe] {
			_, err := validatorFactory.Compile([]byte(r), tpe, nil)
			assert.Error(t, err)
			t.Log(err)
		}
	}
}

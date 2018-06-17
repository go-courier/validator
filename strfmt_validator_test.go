package validator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleNewRegexpStrfmtValidator() {
	AlphaValidator := NewRegexpStrfmtValidator("^[a-zA-Z]+$", "alpha")

	fmt.Println(AlphaValidator.Validate("a"))
	fmt.Println(AlphaValidator.Validate("1"))
	// Output:
	// <nil>
	// invalid alpha
}

func TestStrfmtValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *StrfmtValidator
	}{
		{
			[]interface{}{
				"abc",
				"a",
				"c",
			},
			NewRegexpStrfmtValidator("^[a-zA-Z]+$", "alpha"),
		},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s validate %s", c.validator, v), func(t *testing.T) {
				validator, _ := NewValidatorFactory(c.validator).Compile([]byte("@alpha"), reflect.TypeOf(""), nil)
				err := validator.Validate(v)
				assert.NoError(t, err)
			})
		}
	}
}

func TestStrfmtValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *StrfmtValidator
	}{
		{
			[]interface{}{
				1,
				"1",
				"2",
				"3",
			},
			NewRegexpStrfmtValidator("^[a-zA-Z]+$", "alpha"),
		},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s validate failed %s", c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				assert.Error(t, err)
			})
		}
	}
}

package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorFactory(t *testing.T) {
	f := NewValidatorFactory(BuiltInValidators...)

	cases := []struct {
		rule   string
		values []interface{}
	}{
		{"@int[0,]", []interface{}{1}},
		{"@string[0,]", []interface{}{"1"}},
		{"@slice<@string{1,2}>", []interface{}{
			[]string{},
		}},
	}

	casesForParseFailed := []string{
		"@rule{",
		"@unknown_name",
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s validate %v", c.rule, v), func(t *testing.T) {
				validator, err := f.Compile(c.rule)
				assert.NoError(t, err)
				f.Compile(c.rule)
				assert.NoError(t, validator.Validate(v))
			})
		}
	}

	for _, rule := range casesForParseFailed {
		t.Run("parse failed:"+rule, func(t *testing.T) {
			_, err := f.Compile(rule)
			assert.Error(t, err)
		})
	}
}

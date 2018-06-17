package validator

import (
	"reflect"
	"regexp"
	"fmt"
)

func NewRegexpStrfmtValidator(regexpStr string, name string, aliases ...string) *StrfmtValidator {
	re := regexp.MustCompile(regexpStr)
	validate := func(v interface{}) error {
		if !re.MatchString(v.(string)) {
			return fmt.Errorf("invalid %s", name)
		}
		return nil
	}
	return NewStrfmtValidator(validate, name, aliases...)
}

func NewStrfmtValidator(validate func(v interface{}) error, name string, aliases ... string) *StrfmtValidator {
	return &StrfmtValidator{
		names:    append([]string{name}, aliases...),
		validate: validate,
	}
}

type StrfmtValidator struct {
	names    []string
	validate func(v interface{}) error
}

func (validator *StrfmtValidator) String() string {
	return "@" + validator.names[0]
}

func (validator *StrfmtValidator) Names() []string {
	return validator.names
}

func (validator StrfmtValidator) New(rule *Rule) (Validator, error) {
	return &validator, nil
}

func (validator *StrfmtValidator) Validate(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return NewUnsupportedTypeError("string", reflect.TypeOf(v))
	}
	return validator.validate(s)
}

package validator

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/go-courier/validator/reflectx"
	"github.com/go-courier/validator/rules"
)

func NewValidatorLoader(validatorCreator ValidatorCreator) *ValidatorLoader {
	return &ValidatorLoader{
		ValidatorCreator: validatorCreator,
	}
}

type ValidatorLoader struct {
	ValidatorCreator
	Validator
	PreprocessStage

	Type         reflect.Type
	DefaultValue []byte
	Optional     bool
}

type PreprocessStage int

const (
	PreprocessSkip PreprocessStage = iota
	PreprocessString
	PreprocessPtr
)

func (validator *ValidatorLoader) New(rule *rules.Rule, tpe reflect.Type, validateMgr ValidatorMgr) (Validator, error) {
	loader := NewValidatorLoader(validator.ValidatorCreator)

	loader.Type = tpe
	tpe, loader.PreprocessStage = loader.normalize(tpe)

	v, err := loader.ValidatorCreator.New(rule, tpe, validateMgr)
	if err != nil {
		return nil, err
	}

	loader.Optional = rule.Optional
	loader.DefaultValue = rule.DefaultValue
	loader.Validator = v

	if loader.DefaultValue != nil {
		rv := reflectx.New(loader.Type)
		if err := reflectx.SetValueByString(rv, loader.DefaultValue); err != nil {
			return nil, fmt.Errorf("default value `%s` can not unmarshal to %s: %s", loader.DefaultValue, loader.Type, err)
		}
		if err := loader.Validate(rv); err != nil {
			return nil, fmt.Errorf("default value `%s` is not a valid value of %s: %s", loader.DefaultValue, v, err)
		}
	}

	return loader, nil
}

var interfaceEncodingTextMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

func (ValidatorLoader) normalize(tpe reflect.Type) (reflect.Type, PreprocessStage) {
	if tpe.Implements(interfaceEncodingTextMarshaler) {
		return reflect.TypeOf(""), PreprocessString
	}
	if tpe.Kind() == reflect.Ptr {
		return reflectx.IndirectType(tpe), PreprocessPtr
	}
	return tpe, PreprocessSkip
}

func (validator *ValidatorLoader) Validate(v interface{}) error {
	switch validator.PreprocessStage {
	case PreprocessString:
		if rv, ok := v.(reflect.Value); ok && rv.CanInterface() {
			v = rv.Interface()
		}
		if textMarshaler, ok := v.(encoding.TextMarshaler); ok {
			data, err := textMarshaler.MarshalText()
			if err != nil {
				return err
			}
			v = string(data)
		}
		return validator.Validator.Validate(v)
	default:
		rv, ok := v.(reflect.Value)
		if !ok {
			rv = reflect.ValueOf(&v).Elem()
		}

		isEmptyValue := reflectx.IsEmptyValue(rv)
		if isEmptyValue {
			if !validator.Optional {
				return fmt.Errorf("missing required field")
			}

			if validator.DefaultValue != nil {
				err := reflectx.SetValueByString(rv, validator.DefaultValue)
				if err != nil {
					return fmt.Errorf("unmarshal default value failed")
				}
			}
			return nil
		}

		if rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}
		rv = reflectx.Indirect(rv)
		return validator.Validator.Validate(rv)
	}
}

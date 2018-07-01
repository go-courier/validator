package validator

import (
	"go/ast"
	"reflect"
	"strings"

	"github.com/go-courier/reflectx"

	"github.com/go-courier/validator/errors"
	"github.com/go-courier/validator/rules"
)

func NewStructValidator(namedTagKey string) *StructValidator {
	return &StructValidator{
		namedTagKey:     namedTagKey,
		fieldValidators: map[string]Validator{},
	}
}

type StructValidator struct {
	namedTagKey     string
	Type            reflect.Type
	fieldValidators map[string]Validator
}

func (StructValidator) Names() []string {
	return []string{"struct"}
}

func (validator *StructValidator) Validate(v interface{}) error {
	switch rv := v.(type) {
	case reflect.Value:
		if rv.Type().String() != validator.Type.String() {
			return errors.NewUnsupportedTypeError(rv.Type(), validator.String())
		}
		return validator.ValidateReflectValue(rv)
	default:
		tpe := reflect.TypeOf(v)
		if tpe.String() != validator.Type.String() {
			return errors.NewUnsupportedTypeError(validator.Type, validator.String())
		}
		return validator.ValidateReflectValue(reflect.ValueOf(v))
	}
}

func (validator *StructValidator) ValidateReflectValue(rv reflect.Value) error {
	errSet := errors.NewErrorSet("")
	validator.validate(rv, errSet)
	return errSet.Err()
}

func (validator *StructValidator) validate(rv reflect.Value, errSet *errors.ErrorSet) {
	tpe := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := tpe.Field(i)
		fieldValue := rv.Field(i)
		fieldName, _, exists := fieldInfo(&field, validator.namedTagKey)

		if !ast.IsExported(field.Name) || fieldName == "-" {
			continue
		}

		fieldType := reflectx.IndirectType(field.Type)
		isStructType := fieldType.Kind() == reflect.Struct

		if field.Anonymous && isStructType && !exists {
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				fieldValue = reflectx.New(fieldType)
			}
			validator.validate(fieldValue, errSet)
			continue
		}

		if fieldValidator, ok := validator.fieldValidators[field.Name]; ok {
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				fieldValue = reflectx.New(field.Type)
			}

			err := fieldValidator.Validate(fieldValue)
			errSet.AddErr(err, fieldName)
		}
	}
}

const (
	TagValidate = "validate"
	TagDefault  = "default"
)

func (validator *StructValidator) New(rule *rules.Rule, tpe reflect.Type, mgr ValidatorMgr) (Validator, error) {
	if tpe.Kind() != reflect.Struct {
		return nil, errors.NewUnsupportedTypeError(tpe, validator.String())
	}
	structValidator := NewStructValidator(validator.namedTagKey)
	structValidator.Type = tpe
	errSet := errors.NewErrorSet(tpe.Name())
	structValidator.scan(tpe, errSet, mgr)
	return structValidator, errSet.Err()
}

func (validator *StructValidator) scan(structTpe reflect.Type, errSet *errors.ErrorSet, mgr ValidatorMgr) {
	for i := 0; i < structTpe.NumField(); i++ {
		field := structTpe.Field(i)
		fieldName, omitempty, exists := fieldInfo(&field, validator.namedTagKey)

		if !ast.IsExported(field.Name) || fieldName == "-" {
			continue
		}

		fieldType := reflectx.IndirectType(field.Type)
		isStructType := fieldType.Kind() == reflect.Struct

		if field.Anonymous && isStructType && !exists {
			validator.scan(fieldType, errSet, mgr)
			continue
		}

		tagValidateValue, ok := field.Tag.Lookup(TagValidate)
		if !ok && isStructType {
			tagValidateValue = "@struct"
		}

		fieldValidator, err := mgr.Compile([]byte(tagValidateValue), field.Type, func(rule *rules.Rule) {
			rule.Optional = omitempty

			if defaultValue, ok := field.Tag.Lookup(TagDefault); ok {
				rule.DefaultValue = []byte(defaultValue)
			}
		})
		if err != nil {
			errSet.AddErr(err, field.Name)
			continue
		}
		validator.fieldValidators[field.Name] = fieldValidator
	}
}

func (validator *StructValidator) String() string {
	return "@" + validator.Names()[0]
}

func fieldInfo(structField *reflect.StructField, namedTagKey string) (string, bool, bool) {
	jsonTag, exists := structField.Tag.Lookup(namedTagKey)
	if !exists {
		return structField.Name, false, exists
	}
	omitempty := strings.Index(jsonTag, "omitempty") > 0
	idxOfComma := strings.IndexRune(jsonTag, ',')
	if jsonTag == "" || idxOfComma == 0 {
		return structField.Name, omitempty, true
	}
	if idxOfComma == -1 {
		return jsonTag, omitempty, true
	}
	return jsonTag[0 : idxOfComma-1], omitempty, true
}

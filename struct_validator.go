package validator

import (
	"go/ast"
	"reflect"

	"github.com/go-courier/reflectx"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/validator/errors"
)

func NewStructValidator(namedTagKey string) *StructValidator {
	return &StructValidator{
		namedTagKey:     namedTagKey,
		fieldValidators: map[string]Validator{},
	}
}

type StructValidator struct {
	namedTagKey     string
	fieldValidators map[string]Validator
}

func init() {
	ValidatorMgrDefault.Register(&StructValidator{})
}

func (StructValidator) Names() []string {
	return []string{"struct"}
}

func (validator *StructValidator) Validate(v interface{}) error {
	switch rv := v.(type) {
	case reflect.Value:
		return validator.ValidateReflectValue(rv)
	default:
		return validator.ValidateReflectValue(reflect.ValueOf(v))
	}
}

func (validator *StructValidator) ValidateReflectValue(rv reflect.Value) error {
	errSet := errors.NewErrorSet("")
	validator.validate(rv, errSet)
	return errSet.Err()
}

func (validator *StructValidator) validate(rv reflect.Value, errSet *errors.ErrorSet) {
	typ := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := rv.Field(i)
		fieldName, _, exists := typesutil.FieldDisplayName(field.Tag, validator.namedTagKey, field.Name)

		if !ast.IsExported(field.Name) || fieldName == "-" {
			continue
		}

		fieldType := reflectx.Deref(field.Type)
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

func (validator *StructValidator) New(rule *Rule, mgr ValidatorMgr) (Validator, error) {
	if rule.Type.Kind() != reflect.Struct {
		return nil, errors.NewUnsupportedTypeError(rule.String(), validator.String())
	}

	namedTagKey := ""

	if rule.Rule != nil && len(rule.Params) > 0 {
		namedTagKey = string(rule.Params[0].Bytes())
	}

	structValidator := NewStructValidator(namedTagKey)
	errSet := errors.NewErrorSet("")

	typesutil.EachField(rule.Type, structValidator.namedTagKey, func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		tagValidateValue := field.Tag().Get(TagValidate)

		if tagValidateValue == "" && typesutil.Deref(field.Type()).Kind() == reflect.Struct {
			if _, ok := typesutil.EncodingTextMarshalerTypeReplacer(field.Type()); !ok {
				tagValidateValue = structValidator.String()
			}
		}

		fieldValidator, err := mgr.Compile([]byte(tagValidateValue), field.Type(), func(rule *Rule) {
			if omitempty {
				rule.Optional = omitempty
			}
			if defaultValue, ok := field.Tag().Lookup(TagDefault); ok {
				rule.DefaultValue = []byte(defaultValue)
			}
		})

		if err != nil {
			errSet.AddErr(err, field.Name())
			return true
		}

		if fieldValidator != nil {
			structValidator.fieldValidators[field.Name()] = fieldValidator
		}
		return true
	})

	return structValidator, errSet.Err()
}

func (validator *StructValidator) String() string {
	return "@" + validator.Names()[0] + "<" + validator.namedTagKey + ">"
}

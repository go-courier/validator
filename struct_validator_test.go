package validator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-courier/validator/errors"
)

type SomeTextMarshaler struct {
}

func (*SomeTextMarshaler) MarshalText() ([]byte, error) {
	return []byte("SomeTextMarshaler"), nil
}

func TestStructValidator_New(t *testing.T) {
	type Named string

	type SubPtrStruct struct {
		PtrInt   *int     `validate:"@int[1,]"`
		PtrFloat *float32 `validate:"@float[1,]"`
		PtrUint  *uint    `validate:"@uint[1,]"`
	}

	type SubStruct struct {
		Int   int     `validate:"@int[1,]"`
		Float float32 `validate:"@float[1,]"`
		Uint  uint    `validate:"@uint[1,]"`
	}

	type SomeStruct struct {
		skip         string
		String       string             `json:"String" validate:"@string[1,]"`
		Named        Named              `json:"Named,omitempty" validate:"@string[2,]"`
		PtrString    *string            `json:",omitempty" validate:"@string[3,]"`
		SomeStringer *SomeTextMarshaler `validate:"@string[4,]"`
		Slice        []string           `validate:"@slice<@string[1,]>"`
		SubStruct
		*SubPtrStruct
	}

	validator := NewStructValidator("json")

	validatorFactory := NewValidatorFactory(BuiltInValidators...)

	_, err := validator.New(nil, reflect.TypeOf(&SomeStruct{}).Elem(), validatorFactory)
	assert.NoError(t, err)

	validateStrings := make([]string, 0)

	validatorFactory.cache.Range(func(key, value interface{}) bool {
		validateStrings = append(validateStrings, key.(string))
		return true
	})

	assert.Len(t, validateStrings, 11)
}

func TestStructValidator_NewFailed(t *testing.T) {
	type Named string

	type Struct struct {
		Int    int  `validate:"@int[1,"`
		PtrInt *int `validate:"@uint[2,"`
	}

	type SubStruct struct {
		Float    float32  `validate:"@string"`
		PtrFloat *float32 `validate:"@unknown"`
	}

	type SomeStruct struct {
		skip   string
		String string   `validate:"@string[1,"`
		Named  Named    `validate:"@int"`
		Slice  []string `validate:"@slice<@int>"`
		SubStruct
		Struct Struct
	}

	validator := NewStructValidator("json")

	validatorFactory := NewValidatorFactory(BuiltInValidators...)

	_, err := validator.New(nil, reflect.TypeOf(&SomeStruct{}).Elem(), validatorFactory)
	assert.Error(t, err)
	t.Log(err)

	validateStrings := make([]string, 0)
	validatorFactory.cache.Range(func(key, value interface{}) bool {
		validateStrings = append(validateStrings, key.(string))
		return true
	})
	assert.Len(t, validateStrings, 0)

	{
		_, err := validator.New(nil, reflect.TypeOf(""), validatorFactory)
		assert.Error(t, err)
	}
}

func TestNewStructValidator_Validate(t *testing.T) {
	type Named string

	type SubPtrStruct struct {
		PtrInt   *int     `validate:"@int[1,]"`
		PtrFloat *float32 `validate:"@float[1,]"`
		PtrUint  *uint    `validate:"@uint[1,]"`
	}

	type SubStruct struct {
		Int   int     `validate:"@int[1,]"`
		Float float32 `validate:"@float[1,]"`
		Uint  uint    `validate:"@uint[1,]"`
	}

	type SomeStruct struct {
		skip         string
		String       string             `validate:"@string[1,]"`
		Named        Named              `validate:"@string[2,]"`
		PtrString    *string            `validate:"@string[3,]" default:"123"`
		SomeStringer *SomeTextMarshaler `validate:"@string[20,]"`
		Slice        []string           `validate:"@slice<@string[1,]>"`
		Map          map[string]string  `validate:"@map<@string[2,],@string[1,]>"`
		Struct       SubStruct
		SubStruct
		*SubPtrStruct
	}

	validator := NewStructValidator("json")

	validatorFactory := NewValidatorFactory(BuiltInValidators...)

	structValidator, err := validator.New(nil, reflect.TypeOf(&SomeStruct{}).Elem(), validatorFactory)
	assert.NoError(t, err)

	errForValidate := structValidator.Validate(SomeStruct{
		Slice: []string{"", ""},
		Map: map[string]string{
			"1":  "",
			"11": "",
			"12": "",
		},
	})

	assert.Equal(t, 19, errForValidate.(*errors.ErrorSet).Len())

	t.Log(errForValidate)
}

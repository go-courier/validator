package validator

import (
	"reflect"
	"testing"

	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/validator/errors"
	"github.com/stretchr/testify/require"
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

	ValidatorMgrDefault.ResetCache()

	_, err := validator.New(&Rule{
		Type: typesutil.FromRType(reflect.TypeOf(&SomeStruct{}).Elem()),
	}, ValidatorMgrDefault)
	require.NoError(t, err)

	validateStrings := make([]string, 0)

	ValidatorMgrDefault.cache.Range(func(key, value interface{}) bool {
		validateStrings = append(validateStrings, key.(string))
		return true
	})

	require.Len(t, validateStrings, 11)
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

	ValidatorMgrDefault.ResetCache()
	_, err := validator.New(&Rule{
		Type: typesutil.FromRType(reflect.TypeOf(&SomeStruct{}).Elem()),
	}, ValidatorMgrDefault)
	require.Error(t, err)
	t.Log(err)

	validateStrings := make([]string, 0)

	ValidatorMgrDefault.cache.Range(func(key, value interface{}) bool {
		validateStrings = append(validateStrings, key.(string))
		return true
	})
	require.Len(t, validateStrings, 0)

	{
		_, err := validator.New(&Rule{
			Type: typesutil.FromRType(reflect.TypeOf("")),
		}, ValidatorMgrDefault)
		require.Error(t, err)
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
		JustRequired string
		CanEmpty     *string            `validate:"@string[0,]?"`
		String       string             `validate:"@string[1,]"`
		Named        Named              `validate:"@string[2,]"`
		PtrString    *string            `validate:"@string[3,]" default:"123"`
		SomeStringer *SomeTextMarshaler `validate:"@string[20,]"`
		Slice        []string           `validate:"@slice<@string[1,]>"`
		SliceStruct  []SubStruct        `validate:"@slice"`
		Map          map[string]string  `validate:"@map<@string[2,],@string[1,]>"`
		Struct       SubStruct
		SubStruct
		*SubPtrStruct
	}

	validator := NewStructValidator("json")

	structValidator, err := validator.New(&Rule{
		Type: typesutil.FromRType(reflect.TypeOf(&SomeStruct{}).Elem()),
	}, ValidatorMgrDefault)
	require.NoError(t, err)

	s := SomeStruct{
		Slice: []string{"", ""},
		SliceStruct: []SubStruct{
			{Int: 0},
		},
		Map: map[string]string{
			"1":  "",
			"11": "",
			"12": "",
		},
	}

	errForValidate := structValidator.Validate(s)

	require.Equal(t, 23, errForValidate.(*errors.ErrorSet).Len())
	t.Log(errForValidate)
	require.Nil(t, s.CanEmpty)
}

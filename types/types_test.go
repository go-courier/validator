package types

import (
	"context"
	"encoding"
	"go/types"
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/require"

	"github.com/go-courier/validator/types/__fixtures__/typ"
	typ2 "github.com/go-courier/validator/types/__fixtures__/typ/typ"
)

func TestType(t *testing.T) {
	fn := func(a, b string) bool {
		return true
	}

	values := []interface{}{
		reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem(),
		reflect.TypeOf((*interface {
			encoding.TextMarshaler
			Stringify(ctx context.Context, vs ...interface{}) string
			Add(a, b string) string
			Bytes() []byte
			s() string
		})(nil)).Elem(),

		unsafe.Pointer(t),

		make(typ.Chan),
		make(chan string, 100),

		typ.F,
		typ.Func(fn),
		fn,

		typ.String(""),
		"",
		typ.Bool(true),
		true,
		typ.Int(0),
		ptr.Int(1),
		int(0),
		typ.Int8(0),
		int8(0),
		typ.Int16(0),
		int16(0),
		typ.Int32(0),
		int32(0),
		typ.Int64(0),
		int64(0),
		typ.Uint(0),
		uint(0),
		typ.Uintptr(0),
		uintptr(0),
		typ.Uint8(0),
		uint8(0),
		typ.Uint16(0),
		uint16(0),
		typ.Uint32(0),
		uint32(0),
		typ.Uint64(0),
		uint64(0),
		typ.Float32(0),
		float32(0),
		typ.Float64(0),
		float64(0),
		typ.Complex64(0),
		complex64(0),
		typ.Complex128(0),
		complex128(0),
		typ.Array{},
		[1]string{},
		typ.Slice{},
		[]string{},
		typ.Map{},
		map[string]string{},
		typ.Struct{},
		struct{}{},
		struct {
			typ.Part
			Part2  typ2.Part
			a      string
			A      string `json:"a"`
			Struct struct {
				B string
			}
		}{},
	}

	for i := range values {
		check(t, values[i])
	}
}

func check(t *testing.T, v interface{}) {
	rtype, ok := v.(reflect.Type)
	if !ok {
		rtype = reflect.TypeOf(v)
	}
	ttype := NewTypesTypeFromReflectType(rtype)

	rt := FromRType(rtype)
	tt := FromTType(ttype)

	t.Run(TypeFullName(rt), func(t *testing.T) {
		require.Equal(t, rt.String(), tt.String())
		require.Equal(t, rt.Kind().String(), tt.Kind().String())
		require.Equal(t, rt.Name(), tt.Name())
		require.Equal(t, rt.PkgPath(), tt.PkgPath())
		require.Equal(t, rt.Comparable(), tt.Comparable())
		require.Equal(t, rt.AssignableTo(FromRType(reflect.TypeOf(""))), tt.AssignableTo(FromTType(types.Typ[types.String])))
		require.Equal(t, rt.ConvertibleTo(FromRType(reflect.TypeOf(""))), tt.ConvertibleTo(FromTType(types.Typ[types.String])))

		require.Equal(t, rt.NumMethod(), tt.NumMethod())

		for i := 0; i < rt.NumMethod(); i++ {
			rMethod := rt.Method(i)
			tMethod := tt.Method(i)

			require.Equal(t, rMethod.Name(), tMethod.Name())
			require.Equal(t, rMethod.PkgPath(), tMethod.PkgPath())
			require.Equal(t, rMethod.Type().String(), tMethod.Type().String())
		}

		{
			_, rOk := rt.MethodByName("String")
			_, tOk := tt.MethodByName("String")
			require.Equal(t, rOk, tOk)
		}

		{
			rReplacer, rIs := EncodingTextMarshalerTypeReplacer(rt)
			tReplacer, tIs := EncodingTextMarshalerTypeReplacer(tt)
			require.Equal(t, rIs, tIs)
			require.Equal(t, rReplacer.String(), tReplacer.String())
		}

		if rt.Kind() == reflect.Array {
			require.Equal(t, rt.Len(), tt.Len())
		}

		if rt.Kind() == reflect.Map {
			require.Equal(t, TypeFullName(rt.Key()), TypeFullName(tt.Key()))
		}

		if rt.Kind() == reflect.Array || rt.Kind() == reflect.Slice || rt.Kind() == reflect.Map {
			require.Equal(t, TypeFullName(rt.Elem()), TypeFullName(tt.Elem()))
		}

		if rt.Kind() == reflect.Struct {
			require.Equal(t, rt.NumField(), tt.NumField())

			for i := 0; i < rt.NumField(); i++ {
				rsf := rt.Field(i)
				tsf := tt.Field(i)

				require.Equal(t, rsf.Anonymous(), tsf.Anonymous())
				require.Equal(t, rsf.Tag(), tsf.Tag())
				require.Equal(t, rsf.Name(), tsf.Name())
				require.Equal(t, rsf.PkgPath(), tsf.PkgPath())
				require.Equal(t, TypeFullName(rsf.Type()), TypeFullName(tsf.Type()))
			}

			if rt.NumField() > 0 {
				{
					rsf, _ := rt.FieldByName("A")
					tsf, _ := tt.FieldByName("A")

					require.Equal(t, rsf.Anonymous(), tsf.Anonymous())
					require.Equal(t, rsf.Tag(), tsf.Tag())
					require.Equal(t, rsf.Name(), tsf.Name())
					require.Equal(t, rsf.PkgPath(), tsf.PkgPath())
					require.Equal(t, TypeFullName(rsf.Type()), TypeFullName(tsf.Type()))

					{
						_, ok := rt.FieldByName("_")
						require.False(t, ok)
					}
					{
						_, ok := tt.FieldByName("_")
						require.False(t, ok)
					}
				}

				{
					rsf, _ := rt.FieldByNameFunc(func(s string) bool {
						return s == "A"
					})
					tsf, _ := tt.FieldByNameFunc(func(s string) bool {
						return s == "A"
					})

					require.Equal(t, rsf.Anonymous(), tsf.Anonymous())
					require.Equal(t, rsf.Tag(), tsf.Tag())
					require.Equal(t, rsf.Name(), tsf.Name())
					require.Equal(t, rsf.PkgPath(), tsf.PkgPath())
					require.Equal(t, TypeFullName(rsf.Type()), TypeFullName(tsf.Type()))

					{
						_, ok := rt.FieldByNameFunc(func(s string) bool {
							return false
						})
						require.False(t, ok)
					}
					{
						_, ok := tt.FieldByNameFunc(func(s string) bool {
							return false
						})
						require.False(t, ok)
					}
				}
			}
		}

		if rt.Kind() == reflect.Func {
			require.Equal(t, rt.NumIn(), tt.NumIn())
			require.Equal(t, rt.NumOut(), tt.NumOut())

			for i := 0; i < rt.NumIn(); i++ {
				rParam := rt.In(i)
				tParam := tt.In(i)
				require.Equal(t, rParam.String(), tParam.String())
			}

			for i := 0; i < rt.NumOut(); i++ {
				rResult := rt.Out(i)
				tResult := tt.Out(i)
				require.Equal(t, rResult.String(), tResult.String())
			}
		}

		if rt.Kind() == reflect.Ptr {
			rt = Deref(rt).(*RType)
			tt = Deref(tt).(*TType)

			require.Equal(t, rt.String(), tt.String())
		}
	})
}

func TestTryNew(t *testing.T) {
	{
		_, ok := TryNew(FromRType(reflect.TypeOf(typ.Struct{})))
		require.True(t, ok)
	}
	{
		_, ok := TryNew(FromTType(NewTypesTypeFromReflectType(reflect.TypeOf(typ.Struct{}))))
		require.False(t, ok)
	}
}

func TestEachField(t *testing.T) {
	expect := []string{
		"a", "b", "bool", "c", "Part2",
	}

	{
		rtype := FromRType(reflect.TypeOf(typ.Struct{}))
		names := make([]string, 0)
		EachField(rtype, "json", func(field StructField, fieldDisplayName string, omitempty bool) bool {
			names = append(names, fieldDisplayName)
			return true
		})
		require.Equal(t, expect, names)
	}

	{
		ttype := FromTType(NewTypesTypeFromReflectType(reflect.TypeOf(typ.Struct{})))
		names := make([]string, 0)
		EachField(ttype, "json", func(field StructField, fieldDisplayName string, omitempty bool) bool {
			names = append(names, fieldDisplayName)
			return true
		})
		require.Equal(t, expect, names)
	}
}

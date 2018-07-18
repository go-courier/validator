package types

import (
	"bytes"
	"encoding"
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/go-courier/reflectx"
)

// interface like reflect.Type but only for data type
type Type interface {
	Name() string
	PkgPath() string
	String() string
	Kind() reflect.Kind
	Implements(u Type) bool
	AssignableTo(u Type) bool
	ConvertibleTo(u Type) bool
	Comparable() bool

	Key() Type
	Elem() Type
	Len() int

	NumField() int
	Field(i int) StructField
	FieldByName(name string) (StructField, bool)
	FieldByNameFunc(match func(string) bool) (StructField, bool)

	NumMethod() int
	Method(i int) Method
	MethodByName(name string) (Method, bool)

	NumIn() int
	In(i int) Type
	NumOut() int
	Out(i int) Type
}

type Method interface {
	PkgPath() string
	Name() string
	Type() Type
}

type StructField interface {
	PkgPath() string
	Name() string
	Tag() reflect.StructTag
	Type() Type
	Anonymous() bool
}

func TryNew(u Type) (reflect.Value, bool) {
	switch t := u.(type) {
	case *RType:
		return reflectx.New(t.Type), true
	}
	return reflect.Value{}, false
}

var (
	rtypeEncodingTextMarshaler = FromRType(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem())
	ttypeEncodingTextMarshaler = FromTType(TypeByName("encoding", "TextMarshaler").Underlying())
)

func EncodingTextMarshalerTypeReplacer(u Type) (Type, bool) {
	switch t := u.(type) {
	case *RType:
		return FromRType(reflect.TypeOf("")), t.Implements(rtypeEncodingTextMarshaler)
	case *TType:
		return FromTType(types.Typ[types.String]), t.Implements(ttypeEncodingTextMarshaler)
	}
	return nil, false
}

func EachField(typ Type, tagForName string, each func(field StructField, fieldDisplayName string, omitempty bool) bool) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name()

		fieldDisplayName, omitempty, exists := FieldDisplayName(field.Tag(), tagForName, fieldName)

		if !ast.IsExported(fieldName) || fieldDisplayName == "-" {
			continue
		}

		fieldType := Deref(field.Type())
		isStructType := fieldType.Kind() == reflect.Struct

		if field.Anonymous() && isStructType && !exists {
			EachField(fieldType, tagForName, each)
			continue
		}

		if !each(field, fieldDisplayName, omitempty) {
			break
		}
	}
}

func Deref(typ Type) Type {
	if typ.Kind() == reflect.Ptr {
		return Deref(typ.Elem())
	}
	return typ
}

func TypeFullName(typ Type) string {
	buf := bytes.NewBuffer(nil)

	for typ.Kind() == reflect.Ptr {
		buf.WriteByte('*')
		typ = typ.Elem()
	}

	if name := typ.Name(); name != "" {
		if pkgPath := typ.PkgPath(); pkgPath != "" {
			buf.WriteString(pkgPath)
			buf.WriteRune('.')
		}
		buf.WriteString(name)
		return buf.String()
	}

	buf.WriteString(typ.String())
	return buf.String()
}

func FieldDisplayName(structTag reflect.StructTag, namedTagKey string, defaultName string) (string, bool, bool) {
	jsonTag, exists := structTag.Lookup(namedTagKey)
	if !exists {
		return defaultName, false, exists
	}
	omitempty := strings.Index(jsonTag, "omitempty") > 0
	idxOfComma := strings.IndexRune(jsonTag, ',')
	if jsonTag == "" || idxOfComma == 0 {
		return defaultName, omitempty, true
	}
	if idxOfComma == -1 {
		return jsonTag, omitempty, true
	}
	return jsonTag[0:idxOfComma], omitempty, true
}

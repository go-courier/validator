package validator

import (
	"fmt"
	"reflect"
)

func NewUnsupportedTypeError(need string, tpe reflect.Type) *UnsupportedTypeError {
	return &UnsupportedTypeError{
		Need: need,
		Type: tpe,
	}
}

type UnsupportedTypeError struct {
	Need string
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "unsupported type: " + e.Type.String() + "; need " + e.Need
}

func NewSyntaxErrorf(format string, args ...interface{}) *SyntaxError {
	return &SyntaxError{
		Msg: fmt.Sprintf(format, args...),
	}
}

type SyntaxError struct {
	Msg string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("invalid syntax: %s", e.Msg)
}

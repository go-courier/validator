package errors

import (
	"bytes"
	"reflect"
)

func NewUnsupportedTypeError(tpe reflect.Type, rule string, msgs ...string) *UnsupportedTypeError {
	return &UnsupportedTypeError{
		rule: rule,
		tpe:  tpe,
		msgs: msgs,
	}
}

type UnsupportedTypeError struct {
	msgs []string
	rule string
	tpe  reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.rule)
	buf.WriteString(" could not validate type ")
	buf.WriteString(e.tpe.String())

	for i, msg := range e.msgs {
		if i == 0 {
			buf.WriteString(": ")
		} else {
			buf.WriteString("; ")
		}
		buf.WriteString(msg)
	}

	return buf.String()
}

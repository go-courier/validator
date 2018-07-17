package errors

import (
	"bytes"
)

func NewUnsupportedTypeError(typ string, rule string, msgs ...string) *UnsupportedTypeError {
	return &UnsupportedTypeError{
		rule: rule,
		typ:  typ,
		msgs: msgs,
	}
}

type UnsupportedTypeError struct {
	msgs []string
	rule string
	typ  string
}

func (e *UnsupportedTypeError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.rule)
	buf.WriteString(" could not validate type ")
	buf.WriteString(e.typ)

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

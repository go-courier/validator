package errors

import (
	"fmt"
)

func NewSyntaxError(format string, args ...interface{}) *SyntaxError {
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

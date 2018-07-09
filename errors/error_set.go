package errors

import (
	"bytes"
	"container/list"
	"fmt"
)

func NewErrorSet(root string) *ErrorSet {
	return &ErrorSet{
		root:   root,
		errors: list.New(),
	}
}

type ErrorSet struct {
	root   string
	errors *list.List
}

func (errorSet *ErrorSet) AddErr(err error, keyPathNodes ...interface{}) {
	if err == nil {
		return
	}
	errorSet.errors.PushBack(&FieldError{
		Field: KeyPath(keyPathNodes),
		Error: err,
	})
}

func (errorSet *ErrorSet) Each(cb func(fieldErr *FieldError)) {
	l := errorSet.errors
	for e := l.Front(); e != nil; e = e.Next() {
		if fieldErr, ok := e.Value.(*FieldError); ok {
			cb(fieldErr)
		}
	}
}

func (errorSet *ErrorSet) Flatten() *ErrorSet {
	set := NewErrorSet(errorSet.root)

	errorSet.Each(func(fieldErr *FieldError) {
		if subSet, ok := fieldErr.Error.(*ErrorSet); ok {
			subSet.Flatten().Each(func(subSetFieldErr *FieldError) {
				set.AddErr(subSetFieldErr.Error, append(fieldErr.Field, subSetFieldErr.Field...)...)
			})
		} else {
			set.AddErr(fieldErr.Error, fieldErr.Field...)
		}
	})

	return set
}

func (errorSet *ErrorSet) Len() int {
	return errorSet.Flatten().errors.Len()
}

func (errorSet *ErrorSet) Err() error {
	if errorSet.errors.Len() == 0 {
		return nil
	}
	return errorSet
}

func (errorSet *ErrorSet) Error() string {
	set := errorSet.Flatten()

	buf := bytes.Buffer{}
	set.Each(func(fieldErr *FieldError) {
		buf.WriteString(fmt.Sprintf("%s %s", fieldErr.Field, fieldErr.Error))
		buf.WriteRune('\n')
	})

	return buf.String()
}

type FieldError struct {
	Field KeyPath
	Error error `json:"msg"`
}

type KeyPath []interface{}

func (keyPath KeyPath) String() string {
	buf := &bytes.Buffer{}
	for i := 0; i < len(keyPath); i++ {
		switch keyOrIndex := keyPath[i].(type) {
		case string:
			if buf.Len() > 0 {
				buf.WriteRune('.')
			}
			buf.WriteString(keyOrIndex)
		case int:
			buf.WriteString(fmt.Sprintf("[%d]", keyOrIndex))
		}
	}
	return buf.String()
}

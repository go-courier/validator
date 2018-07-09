package errors

import (
	"fmt"
	"reflect"
)

func ExampleUnsupportedTypeError() {
	fmt.Println(NewUnsupportedTypeError(reflect.TypeOf(""), "@int", "something wrong", "something wrong"))
	// Output:
	// @int could not validate type string: something wrong; something wrong
}

package errors

import (
	"fmt"
)

func ExampleUnsupportedTypeError() {
	fmt.Println(NewUnsupportedTypeError("string", "@int", "something wrong", "something wrong"))
	// Output:
	// @int could not validate type string: something wrong; something wrong
}

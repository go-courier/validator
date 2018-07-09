package errors

import (
	"fmt"
)

func ExampleSyntaxError() {
	fmt.Println(NewSyntaxError("rule"))
	// Output:
	// invalid syntax: rule
}

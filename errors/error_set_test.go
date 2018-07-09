package errors

import (
	"fmt"
)

func ExampleErrorSet() {
	subErrSet := NewErrorSet("")
	subErrSet.AddErr(fmt.Errorf("err"), "PropA")
	subErrSet.AddErr(fmt.Errorf("err"), "PropB")

	errSet := NewErrorSet("")
	errSet.AddErr(fmt.Errorf("err"), "Key")
	errSet.AddErr(subErrSet.Err(), "Key", 1)
	errSet.AddErr(NewErrorSet("").Err(), "Key", 1)

	fmt.Println(errSet.Len())
	fmt.Println(errSet)
	// Output:
	// 3
	// Key err
	// Key[1].PropA err
	// Key[1].PropB err
}

package typ

import (
	"github.com/go-courier/validator/types/__fixtures__/typ/typ"
)

type String string
type Bool bool
type Int int
type Int8 int8
type Int16 int16
type Int32 int32
type Int64 int64
type Uint uint
type Uint8 uint8
type Uint16 uint16
type Uint32 uint32
type Uint64 uint64
type Uintptr uintptr
type Float32 float32
type Float64 float64
type Complex64 complex64
type Complex128 complex128

type Array [1]string

type Map map[string]string
type Slice []string
type Chan chan string
type Func func(a, b string) bool

func F() {}

type Struct struct {
	a    string
	A    string `json:"a"`
	B    string `json:"b"`
	Bool `json:"bool,omitempty"`
	typ.Part
	Part2 Part `json:",omitempty"`
}

func (Struct) String() string {
	return ""
}

type Part struct {
	C string `json:"c"`
}

type Interface interface {
	String() string
}

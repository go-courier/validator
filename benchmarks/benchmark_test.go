package benchmarks

import (
	"encoding"
	"reflect"
	"testing"
)

type SomeString string

func (SomeString) AnonymousRecv() {
}

func (v SomeString) Recv() {

}

func (v *SomeString) PtrRecv() {

}

type SomeStruct struct {
	// nolint: unused
	a, b, c, d, e, f, g, h, i, j, k, l, m, n string
}

func (SomeStruct) AnonymousRecv() {
}

func (v SomeStruct) Recv() {

}

func (v *SomeStruct) PtrRecv() {

}

type SomeSlice []int

func (SomeSlice) AnonymousRecv() {
}

func (v SomeSlice) Recv() {

}

func (v *SomeSlice) PtrRecv() {

}

func BenchmarkMethodPerformance(b *testing.B) {
	b.Run("string", func(b *testing.B) {
		val := SomeString("")

		for i := 0; i < b.N; i++ {
			_ = val
		}
	})

	b.Run("string.AnonymousRecv()", func(b *testing.B) {
		val := SomeString("")

		for i := 0; i < b.N; i++ {
			val.AnonymousRecv()
		}
	})

	b.Run("string.Recv()", func(b *testing.B) {
		val := SomeString("")

		for i := 0; i < b.N; i++ {
			val.Recv()
		}
	})

	b.Run("struct.PtrRecv()", func(b *testing.B) {
		val := SomeString("")

		for i := 0; i < b.N; i++ {
			val.PtrRecv()
		}
	})

	b.Run("struct.x", func(b *testing.B) {
		val := SomeStruct{}

		for i := 0; i < b.N; i++ {
			_ = val.a
		}
	})

	b.Run("struct.AnonymousRecv()", func(b *testing.B) {
		val := SomeStruct{}

		for i := 0; i < b.N; i++ {
			val.AnonymousRecv()
		}
	})

	b.Run("struct.Recv()", func(b *testing.B) {
		val := SomeStruct{}

		for i := 0; i < b.N; i++ {
			val.Recv()
		}
	})

	b.Run("struct.PtrRecv()", func(b *testing.B) {
		val := SomeStruct{}

		for i := 0; i < b.N; i++ {
			val.PtrRecv()
		}
	})

	b.Run("slice.x", func(b *testing.B) {
		val := SomeSlice{}

		for i := 0; i < b.N; i++ {
			_ = val
		}
	})

	b.Run("slice.AnonymousRecv()", func(b *testing.B) {
		val := SomeSlice{}

		for i := 0; i < b.N; i++ {
			val.AnonymousRecv()
		}
	})

	b.Run("slice.Recv()", func(b *testing.B) {
		val := SomeSlice{}

		for i := 0; i < b.N; i++ {
			val.Recv()
		}
	})

	b.Run("slice.PtrRecv()", func(b *testing.B) {
		val := SomeSlice{}

		for i := 0; i < b.N; i++ {
			val.PtrRecv()
		}
	})
}

var (
	textMarshalerType = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
)

func BenchmarkImplementsOrTypeAssign(b *testing.B) {
	val := 1
	rv := reflect.ValueOf(&val)

	b.Run("implements", func(b *testing.B) {
		typ := rv.Type()

		for i := 0; i < b.N; i++ {
			_ = typ.Implements(textMarshalerType)
		}
	})

	b.Run("interface and type assert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rv.Interface().(encoding.TextMarshaler)
		}
	})

	b.Run("type assert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = (interface{})(val).(encoding.TextMarshaler)
		}
	})
}

func BenchmarkValue(b *testing.B) {
	val := 1
	rv := reflect.ValueOf(&val)

	b.Run("use directly", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = val
		}
	})

	b.Run("type assertion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = (interface{})(val).(int)
		}
	})

	b.Run("reflect.TypeOf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = reflect.TypeOf(val)
		}
	})

	b.Run("reflect.Value.Interface", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rv.Interface()
		}
	})

	b.Run("reflect.Value.CanInterface", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rv.CanInterface()
		}
	})

	b.Run("reflect.Value.Interface then type assertion", func(b *testing.B) {
		v := rv.Interface()
		for i := 0; i < b.N; i++ {
			_ = v.(*int)
		}
	})

	b.Run("reflect.Value.Elem.Interface and type assertion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rv.Elem().Interface().(int)
		}
	})

	b.Run("reflect.ValueOf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = reflect.ValueOf(val)
		}
	})
}

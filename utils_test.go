package validator

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/assert"
)

type Zero string

func (Zero) IsZero() bool {
	return true
}

func TestIsEmptyValue(t *testing.T) {

	emptyValues := []interface{}{
		Zero(""),
		(*string)(nil),
		(interface{})(nil),
		(struct {
			V interface{}
		}{}).V,
		"",
		0,
		uint(0),
		float32(0),
		false,
	}
	for _, v := range emptyValues {
		assert.True(t, IsEmptyValue(reflect.ValueOf(v)))
	}
}

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(data []byte) error {
	dur, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

func TestUnmarshalDefaultValue(t *testing.T) {
	v := struct {
		Duration    Duration
		PtrDuration *Duration
		String      string
		PtrString   *string
		Int         int
		PtrInt      *int
		Uint        uint
		PtrUint     *uint
		Float       float32
		PtrFloat    *float32
		Bool        bool
		PtrBool     *bool
	}{}

	rv := reflect.ValueOf(&v).Elem()

	d := Duration(2 * time.Second)

	cases := []struct {
		rv           reflect.Value
		defaultValue string
		expect       interface{}
	}{
		{
			rv.FieldByName("PtrString"),
			"",
			ptr.String(""),
		},
		{
			rv.FieldByName("Duration"),
			"2s",
			Duration(2 * time.Second),
		},
		{
			rv.FieldByName("PtrDuration"),
			"2s",
			&d,
		},
		{
			rv.FieldByName("String"),
			"string",
			"string",
		},
		{
			rv.FieldByName("PtrString"),
			"string",
			ptr.String("string"),
		},
		{
			rv.FieldByName("Int"),
			"1",
			1,
		},
		{
			rv.FieldByName("PtrInt"),
			"1",
			ptr.Int(1),
		},
		{
			rv.FieldByName("Uint"),
			"1",
			uint(1),
		},
		{
			rv.FieldByName("PtrUint"),
			"1",
			ptr.Uint(1),
		},
		{
			rv.FieldByName("Float"),
			"1",
			float32(1),
		},
		{
			rv.FieldByName("PtrFloat"),
			"1",
			ptr.Float32(1),
		},
		{
			rv.FieldByName("Bool"),
			"true",
			true,
		},
		{
			rv.FieldByName("PtrBool"),
			"true",
			ptr.Bool(true),
		},
	}
	for _, c := range cases {
		err := UnmarshalDefaultValue(c.rv, []byte(c.defaultValue))
		assert.NoError(t, err)
		assert.Equal(t, c.expect, c.rv.Interface())
	}
}

func TestMaxIntAndMinInt(t *testing.T) {
	cases := [][]int64{
		{MinInt(8), -128},
		{MaxInt(8), 127},
		{MinInt(16), -32768},
		{MaxInt(16), 32767},
		{MinInt(32), -2147483648},
		{MaxInt(32), 2147483647},
		{MinInt(64), -9223372036854775808},
		{MaxInt(64), 9223372036854775807},
	}
	for _, values := range cases {
		assert.Equal(t, values[1], values[0])
	}
}

package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldError(t *testing.T) {
	data, _ := json.Marshal(FieldError{
		Field: KeyPath{
			"key",
			1,
			"string",
		},
		Msg: "error",
	})

	assert.Equal(t, `{"field":"key[1].string","msg":"error"}`, string(data))
}

package customtype_test

import (
	"encoding/json"
	"testing"

	"github.com/Medzoner/medzoner-go/internal/domain/customtype"
	"gotest.tools/assert"
)

func TestNullString(t *testing.T) {
	t.Run("Unit: test NullString with valid value", func(t *testing.T) {
		ns := customtype.NullString{String: "hello", Valid: true}
		assert.Equal(t, ns.String, "hello")
		assert.Equal(t, ns.Valid, true)
	})

	t.Run("Unit: test NullString with empty value", func(t *testing.T) {
		ns := customtype.NullString{}
		assert.Equal(t, ns.String, "")
		assert.Equal(t, ns.Valid, false)
	})
}

func TestNewNullString(t *testing.T) {
	t.Run("Unit: test NewNullString creates valid NullString", func(t *testing.T) {
		ns := customtype.NewNullString("test")
		assert.Equal(t, ns.String, "test")
		assert.Equal(t, ns.Valid, true)
	})
}

func TestNullString_Scan(t *testing.T) {
	t.Run("Unit: test Scan with string value", func(t *testing.T) {
		var ns customtype.NullString
		err := ns.Scan("hello")
		assert.NilError(t, err)
		assert.Equal(t, ns.String, "hello")
		assert.Equal(t, ns.Valid, true)
	})

	t.Run("Unit: test Scan with []byte value", func(t *testing.T) {
		var ns customtype.NullString
		err := ns.Scan([]byte("world"))
		assert.NilError(t, err)
		assert.Equal(t, ns.String, "world")
		assert.Equal(t, ns.Valid, true)
	})

	t.Run("Unit: test Scan with nil value", func(t *testing.T) {
		ns := customtype.NullString{String: "previous", Valid: true}
		err := ns.Scan(nil)
		assert.NilError(t, err)
		assert.Equal(t, ns.String, "")
		assert.Equal(t, ns.Valid, false)
	})

	t.Run("Unit: test Scan with unsupported type", func(t *testing.T) {
		var ns customtype.NullString
		err := ns.Scan(12345)
		assert.NilError(t, err)
		assert.Equal(t, ns.String, "")
		assert.Equal(t, ns.Valid, false)
	})
}

func TestNullString_Value(t *testing.T) {
	t.Run("Unit: test Value with valid NullString", func(t *testing.T) {
		ns := customtype.NullString{String: "hello", Valid: true}
		val, err := ns.Value()
		assert.NilError(t, err)
		assert.Equal(t, val, "hello")
	})

	t.Run("Unit: test Value with invalid NullString returns nil", func(t *testing.T) {
		ns := customtype.NullString{Valid: false}
		val, err := ns.Value()
		assert.NilError(t, err)
		assert.Assert(t, val == nil)
	})
}

func TestNullString_MarshalJSON(t *testing.T) {
	t.Run("Unit: test MarshalJSON with valid NullString", func(t *testing.T) {
		ns := customtype.NullString{String: "hello", Valid: true}
		data, err := json.Marshal(ns)
		assert.NilError(t, err)
		assert.Equal(t, string(data), `"hello"`)
	})

	t.Run("Unit: test MarshalJSON with invalid NullString returns null", func(t *testing.T) {
		ns := customtype.NullString{Valid: false}
		data, err := json.Marshal(ns)
		assert.NilError(t, err)
		assert.Equal(t, string(data), `null`)
	})
}

func TestNullString_UnmarshalJSON(t *testing.T) {
	t.Run("Unit: test UnmarshalJSON with string value", func(t *testing.T) {
		var ns customtype.NullString
		err := json.Unmarshal([]byte(`"world"`), &ns)
		assert.NilError(t, err)
		assert.Equal(t, ns.String, "world")
		assert.Equal(t, ns.Valid, true)
	})

	t.Run("Unit: test UnmarshalJSON with null value", func(t *testing.T) {
		ns := customtype.NullString{String: "previous", Valid: true}
		err := json.Unmarshal([]byte(`null`), &ns)
		assert.NilError(t, err)
		assert.Equal(t, ns.String, "")
		assert.Equal(t, ns.Valid, false)
	})

	t.Run("Unit: test UnmarshalJSON with invalid JSON", func(t *testing.T) {
		var ns customtype.NullString
		err := json.Unmarshal([]byte(`{invalid`), &ns)
		assert.Assert(t, err != nil)
	})
}

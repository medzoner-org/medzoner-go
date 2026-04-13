package customtype_test

import (
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


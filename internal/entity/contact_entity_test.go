package entity_test

import (
	"testing"
	"time"

	"github.com/Medzoner/medzoner-go/internal/domain/customtype"
	"github.com/Medzoner/medzoner-go/internal/entity"
	"gotest.tools/assert"
)

func TestContact_Get(t *testing.T) {
	t.Run("Unit: test Contact Get returns message", func(t *testing.T) {
		contact := &entity.Contact{
			Name:    "John",
			Message: "Hello World",
			Email:   customtype.NullString{String: "john@example.com", Valid: true},
			DateAdd: time.Now(),
			UUID:    "test-uuid",
		}

		result := contact.Get()

		assert.Equal(t, result, "Hello World")
	})

	t.Run("Unit: test Contact Get returns empty message", func(t *testing.T) {
		contact := &entity.Contact{}

		result := contact.Get()

		assert.Equal(t, result, "")
	})
}


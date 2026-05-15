package entity_test

import (
	"testing"
	"time"

	"github.com/Medzoner/medzoner-go/internal/entity"
	"gopkg.in/guregu/null.v1"
	"gotest.tools/assert"
)

func TestContact_Get(t *testing.T) {
	t.Run("Unit: test Contact Get returns message", func(t *testing.T) {
		contact := &entity.Contact{
			Name:    "John",
			Message: "Hello World",
			Email:   null.StringFrom("john@example.com"),
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

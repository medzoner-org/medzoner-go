package domains

import (
	"time"

	"github.com/google/uuid"
	"gopkg.in/guregu/null.v1"
)

type Contact struct {
	DateAdd time.Time   `db:"date_add"`
	UUID    uuid.UUID   `db:"uuid"    json:"uuid"`
	Name    string      `db:"name"`
	Message string      `db:"message"`
	Email   null.String `db:"email"`
	ID      int         `db:"id"      json:"id"`
}

func (c *Contact) Get() string {
	return c.Message
}

func (c *Contact) Template() string {
	return "/tmpl/contact/contactEmail.html"
}

func (c *Contact) EmailValue() string {
	return c.Email.String
}

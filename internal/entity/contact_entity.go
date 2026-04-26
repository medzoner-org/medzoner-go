package entity

import (
	"time"

	"github.com/Medzoner/medzoner-go/internal/domain/customtype"
)

// Contact represents a contact form submission.
type Contact struct {
	DateAdd time.Time             `db:"date_add"`
	UUID    string                `db:"uuid"    json:"uuid"`
	Name    string                `db:"name"`
	Message string                `db:"message"`
	Email   customtype.NullString `db:"email"`
	ID      int                   `db:"id"      json:"id"`
}

// Get returns the contact message.
func (c *Contact) Get() string {
	return c.Message
}

func (c *Contact) Template() string {
	return "/tmpl/contact/contactEmail.html"
}

// EmailValue returns the email string value for use in queries.
func (c *Contact) EmailValue() string {
	return c.Email.String
}

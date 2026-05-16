package event

import (
	"github.com/Medzoner/medzoner-go/internal/domains"
)

// ContactCreatedEvent is a struct that implements Event interface and contains model Contact
type ContactCreatedEvent struct {
	Contact domains.Contact
}

// GetModel returns model Contact
func (c ContactCreatedEvent) GetModel() any {
	return c.Contact
}

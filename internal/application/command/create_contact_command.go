package command

import "time"

// CreateContactCommand is a struct that contains contact data
type CreateContactCommand struct {
	DateAdd time.Time `json:"dateAdd" validate:"required"`
	Name    string    `json:"name"    validate:"required,min=3,max=255"`
	Email   string    `json:"email"   validate:"required,email,min=3,max=255"`
	Message string    `json:"message" validate:"required,min=5,max=1500"`
}

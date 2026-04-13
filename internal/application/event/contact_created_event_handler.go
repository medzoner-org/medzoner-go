package event

import (
	"context"
	"fmt"

	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/application/service/mailer"
	"github.com/Medzoner/medzoner-go/internal/entity"
)

var (
	// ErrEmptyUUID is returned when a contact has an empty UUID.
	ErrEmptyUUID = fmt.Errorf("contact UUID is empty")
	// ErrInvalidEventType is returned when the event model is not of type entity.Contact.
	ErrInvalidEventType = fmt.Errorf("contact is not of type entity.Contact")
)

// ContactCreatedEventHandler handles ContactCreatedEvent and sends a notification mail.
type ContactCreatedEventHandler struct {
	Mailer mailer.Mailer
}

// NewContactCreatedEventHandler creates a new ContactCreatedEventHandler.
func NewContactCreatedEventHandler(mailer mailer.Mailer) *ContactCreatedEventHandler {
	return &ContactCreatedEventHandler{
		Mailer: mailer,
	}
}

// Publish handles event ContactCreatedEvent and sends mail to admin.
func (c ContactCreatedEventHandler) Publish(ctx context.Context, event Event) error {
	ctx, iSpan := observability.StartSpan(ctx, "ContactCreatedEventHandler.Publish")
	defer iSpan.End()

	if _, ok := event.(ContactCreatedEvent); !ok {
		return nil
	}

	contact, ok := event.GetModel().(entity.Contact)
	if !ok {
		return fmt.Errorf("error during get contact from event: %w", ErrInvalidEventType)
	}

	if contact.UUID == "" {
		return fmt.Errorf("error during get contact from event: %w", ErrEmptyUUID)
	}

	if _, err := c.Mailer.Send(ctx, &contact); err != nil {
		return fmt.Errorf("error during send mail: %w", err)
	}

	return nil
}

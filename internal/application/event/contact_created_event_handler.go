package event

import (
	"context"
	"fmt"

	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/application/service/mailer"
	"github.com/Medzoner/medzoner-go/internal/entity"
)

// ContactCreatedEventHandler is a struct that implements EventHandler interface and handle ContactCreatedEvent
type ContactCreatedEventHandler struct {
	Mailer mailer.Mailer
}

// NewContactCreatedEventHandler is a function that returns a new ContactCreatedEventHandler
func NewContactCreatedEventHandler(mailer mailer.Mailer) *ContactCreatedEventHandler {
	return &ContactCreatedEventHandler{
		Mailer: mailer,
	}
}

// Publish handles event ContactCreatedEvent and send mail to admin
func (c ContactCreatedEventHandler) Publish(ctx context.Context, event Event) error {
	ctx, iSpan := observability.StartSpan(ctx, "ContactCreatedEventHandler.Publish")
	defer iSpan.End()

	switch event.(type) {
	case ContactCreatedEvent:
		if contactCreated, ok := event.GetModel().(entity.Contact); ok {
			if contactCreated.UUID == "" {
				return fmt.Errorf("error during get contact from event: %w", fmt.Errorf("contact UUID is empty"))
			}
			_, err := c.Mailer.Send(ctx, &contactCreated)
			if err != nil {
				//return fmt.Errorf("error during send mail: %w", c.Telemetry.ErrorSpan(iSpan, err))
				return fmt.Errorf("error during send mail: %w", err)
			}
			//c.Telemetry.Log(ctx, "Mail was send.")
			return nil
		}
		return fmt.Errorf("error during get contact from event: %w", fmt.Errorf("contact is not of type entity.Contact"))
	default:
		//c.Telemetry.Error(ctx, "ErrorSpan bad event type.")
	}

	return nil
}

package command

import (
	"context"
	"fmt"
	"time"

	"github.com/Medzoner/gomedz/pkg/observability"
	event2 "github.com/Medzoner/medzoner-go/internal/application/event"
	"github.com/Medzoner/medzoner-go/internal/domains"
	"github.com/Medzoner/medzoner-go/internal/ports/contact"
	"github.com/docker/distribution/uuid"
	"gopkg.in/guregu/null.v1"
)

// CreateContactCommandHandler is a struct that implements CommandHandler interface and handle CreateContactCommand
type CreateContactCommandHandler struct {
	ContactRepository          contact.Repository
	ContactCreatedEventHandler event2.Handler
}

// NewCreateContactCommandHandler is a function that returns a new CreateContactCommandHandler
func NewCreateContactCommandHandler(
	contactRepository contact.Repository,
	contactCreatedEventHandler event2.Handler,
) CreateContactCommandHandler {
	return CreateContactCommandHandler{
		ContactRepository:          contactRepository,
		ContactCreatedEventHandler: contactCreatedEventHandler,
	}
}

// Handle handles command CreateContactCommand and create contact in database and send mail to admin with event ContactCreatedEvent
func (c *CreateContactCommandHandler) Handle(ctx context.Context, command CreateContactCommand) error {
	ctx, iSpan := observability.StartSpan(ctx, "CreateContactCommandHandler.Publish")
	defer iSpan.End()

	ct := domains.Contact{
		Name:    command.Name,
		Message: command.Message,
		Email:   null.StringFrom(command.Email),
		DateAdd: time.Now(),
		UUID:    uuid.UUID{}.String(),
	}
	if err := c.ContactRepository.Save(ctx, ct); err != nil {
		return fmt.Errorf("error during save contact: %w", err)
	}

	if err := c.ContactCreatedEventHandler.Publish(ctx, event2.ContactCreatedEvent{Contact: ct}); err != nil {
		return fmt.Errorf("error during handle event: %w", err)
	}

	return nil
}

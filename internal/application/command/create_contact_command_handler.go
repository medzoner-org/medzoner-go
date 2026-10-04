package command

import (
	"context"
	"fmt"
	"time"

	event2 "github.com/Medzoner/medzoner-go/internal/application/event"
	"github.com/Medzoner/medzoner-go/internal/domains"
	"github.com/Medzoner/medzoner-go/internal/ports/contact"
	"github.com/google/uuid"
	"github.com/medzoner-org/gomedz/pkg/observability"
	"gopkg.in/guregu/null.v1"
)

// CreateContactCommandHandler is a struct that implements CommandHandler interface and handle CreateContactCommand
type CreateContactCommandHandler struct {
	ContactRepository          contact.Repository
	ContactCreatedEventHandler event2.Handler
}

// NewCreateContactHandler is a function that returns a new CreateContactCommandHandler
func NewCreateContactHandler(
	contactRepository contact.Repository,
	contactCreatedEventHandler event2.Handler,
) CreateContactCommandHandler {
	return CreateContactCommandHandler{
		ContactRepository:          contactRepository,
		ContactCreatedEventHandler: contactCreatedEventHandler,
	}
}

// Handle handles command CreateContactCommand and create contact in database and send mail to admin with event ContactCreatedEvent
func (c *CreateContactCommandHandler) Handle(ctx context.Context, cmd CreateContactCommand) error {
	ctx, span := observability.StartSpan(ctx, "CreateContactCommandHandler.Publish")
	defer span.End()

	ct := domains.Contact{
		Name:    cmd.Name,
		Message: cmd.Message,
		Email:   null.StringFrom(cmd.Email),
		DateAdd: time.Now(),
		UUID:    uuid.New(),
	}
	if err := c.ContactRepository.Save(ctx, ct); err != nil {
		return fmt.Errorf("error during save contact: %w", err)
	}

	if err := c.ContactCreatedEventHandler.Publish(ctx, event2.ContactCreatedEvent{Contact: ct}); err != nil {
		return fmt.Errorf("error during handle event: %w", err)
	}

	return nil
}

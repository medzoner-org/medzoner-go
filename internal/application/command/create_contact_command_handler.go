package command

import (
	"context"
	"fmt"
	"time"

	event2 "github.com/Medzoner/medzoner-go/internal/application/event"
	"github.com/Medzoner/medzoner-go/internal/domain/repository"

	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/entity"
	"github.com/docker/distribution/uuid"
	"gopkg.in/guregu/null.v1"
)

// CreateContactCommandHandler is a struct that implements CommandHandler interface and handle CreateContactCommand
type CreateContactCommandHandler struct {
	ContactRepository          repository.ContactRepository
	ContactCreatedEventHandler event2.IEventHandler
}

// NewCreateContactCommandHandler is a function that returns a new CreateContactCommandHandler
func NewCreateContactCommandHandler(
	contactRepository repository.ContactRepository,
	contactCreatedEventHandler event2.IEventHandler,
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

	contact := entity.Contact{
		Name:    command.Name,
		Message: command.Message,
		Email:   null.StringFrom(command.Email),
		DateAdd: time.Now(),
		UUID:    uuid.UUID{}.String(),
	}
	if err := c.ContactRepository.Save(ctx, contact); err != nil {
		return fmt.Errorf("error during save contact: %w", err)
	}

	if err := c.ContactCreatedEventHandler.Publish(ctx, event2.ContactCreatedEvent{Contact: contact}); err != nil {
		return fmt.Errorf("error during handle event: %w", err)
	}

	return nil
}

package command_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	command2 "github.com/Medzoner/medzoner-go/internal/application/command"
	"github.com/Medzoner/medzoner-go/internal/application/event"

	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/domains"
	mocks "github.com/Medzoner/medzoner-go/test"
	"github.com/golang/mock/gomock"
	"gotest.tools/assert"
)

func init() {
	l, err := logger.NewLogger(logger.Config{Level: "debug"})
	if err != nil {
		panic(err)
	}
	_, _ = observability.NewTelemetry(context.Background(), observability.Config{}, l)
}

func TestCreateContactCommandHandler(t *testing.T) {
	t.Run("Unit: test CreateContactCommandHandler success", func(t *testing.T) {
		date := time.Time{}
		createContactCommand := command2.CreateContactCommand{
			Name:    "a name",
			Email:   "an email",
			Message: "the message",
			DateAdd: date,
		}

		mocked := mocks.New(t)
		mocked.Mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

		handler := command2.NewCreateContactCommandHandler(
			&ContactRepositoryTest{}, event.ContactCreatedEventHandler{Mailer: mocked.Mailer},
		)

		err := handler.Handle(context.Background(), createContactCommand)

		assert.Equal(t, err, nil)
	})
	t.Run("Unit: test CreateContactCommandHandler error save db", func(t *testing.T) {
		date := time.Time{}
		createContactCommand := command2.CreateContactCommand{
			Name:    "a name",
			Email:   "email@example.com",
			Message: "the message",
			DateAdd: date,
		}

		mocked := mocks.New(t)
		mocked.ContactRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

		handler := command2.NewCreateContactCommandHandler(
			mocked.ContactRepository, event.ContactCreatedEventHandler{
				Mailer: mocked.Mailer,
			},
		)
		err := handler.Handle(context.Background(), createContactCommand)

		assert.Error(t, err, "error during save contact: error")
	})
	t.Run("Unit: test CreateContactCommandHandler error send mail", func(t *testing.T) {
		date := time.Time{}
		createContactCommand := command2.CreateContactCommand{
			Name:    "a name",
			Email:   "email@example.com",
			Message: "the message",
			DateAdd: date,
		}

		mocked := mocks.New(t)
		mocked.Mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(false, errors.New("error")).Times(1)

		handler := command2.NewCreateContactCommandHandler(
			&ContactRepositoryTest{}, event.ContactCreatedEventHandler{Mailer: mocked.Mailer},
		)

		err := handler.Handle(context.Background(), createContactCommand)

		assert.Error(t, err, "error during handle event: error during send mail: error")
	})
}

type ContactRepositoryTest struct{}

func (r ContactRepositoryTest) Save(ctx context.Context, contact domains.Contact) error {
	_ = ctx
	fmt.Println(contact)
	return nil
}

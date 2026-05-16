package event_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	event2 "github.com/Medzoner/medzoner-go/internal/application/event"
	mocks "github.com/Medzoner/medzoner-go/test"
	"github.com/golang/mock/gomock"

	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/domains"
	"gopkg.in/guregu/null.v1"
	"gotest.tools/assert"
)

func init() {
	l, err := logger.NewLogger(logger.Config{Level: "debug"})
	if err != nil {
		panic(err)
	}
	_, _ = observability.NewTelemetry(context.Background(), observability.Config{}, l)
}

func TestContactCreatedEventHandler(t *testing.T) {
	contact := domains.Contact{
		Name:    "a name",
		Email:   null.StringFrom("an email"),
		Message: "the message",
		DateAdd: time.Time{},
		ID:      1,
		UUID:    "a uuid",
	}

	t.Run("Unit: test ContactCreatedEventHandler success", func(t *testing.T) {
		mocked := mocks.New(t)
		mailer := mocked.Mailer
		mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

		contactCreatedEvent := event2.ContactCreatedEvent{
			Contact: contact,
		}

		handler := event2.ContactCreatedEventHandler{
			Mailer: mailer,
		}

		err := handler.Publish(context.Background(), contactCreatedEvent)
		assert.Equal(t, err, nil)
	})
	t.Run("Unit: test ContactCreatedEventHandler failed with bad event", func(t *testing.T) {
		mocked := mocks.New(t)
		mailer := mocked.Mailer

		handler := event2.NewContactCreatedEventHandler(mailer)

		err := handler.Publish(context.Background(), BadEvent{})
		assert.Equal(t, err, nil)
	})
	t.Run("Unit: test ContactCreatedEventHandler failed send mail", func(t *testing.T) {
		mocked := mocks.New(t)
		mailer := mocked.Mailer
		mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("error")).AnyTimes()

		handler := event2.NewContactCreatedEventHandler(mailer)
		contactCreatedEvent := event2.ContactCreatedEvent{
			Contact: contact,
		}
		err := handler.Publish(context.Background(), contactCreatedEvent)
		assert.Error(t, err, "error during send mail: error")
	})
	t.Run("Unit: test ContactCreatedEventHandler failed with empty UUID", func(t *testing.T) {
		mocked := mocks.New(t)
		mailer := mocked.Mailer

		handler := event2.NewContactCreatedEventHandler(mailer)
		contactCreatedEvent := event2.ContactCreatedEvent{
			Contact: domains.Contact{
				Name:    "test",
				Email:   null.StringFrom("test@test.com"),
				Message: "msg",
				DateAdd: time.Time{},
				UUID:    "",
			},
		}
		err := handler.Publish(context.Background(), contactCreatedEvent)
		assert.ErrorContains(t, err, "contact UUID is empty")
	})
}

type BadEvent struct{}

func (b BadEvent) GetName() string {
	return "BadEvent"
}

func (b BadEvent) GetModel() any {
	return BadEvent{}
}

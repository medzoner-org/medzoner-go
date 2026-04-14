package query_test

import (
	"context"
	"errors"
	"testing"

	query2 "github.com/Medzoner/medzoner-go/internal/application/query"
	mocks "github.com/Medzoner/medzoner-go/test"

	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
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

func TestListTechnoQueryHandler(t *testing.T) {
	t.Run("Unit: test ListTechnoQueryHandler \"stack\" success", func(t *testing.T) {
		listTechnoQuery := query2.ListTechnoQuery{
			Type: "stack",
		}

		handler := query2.NewListTechnoQueryHandler(&TechnoRepositoryTest{})

		_, err := handler.Handle(context.Background(), listTechnoQuery)
		if err != nil {
			t.Errorf("ErrorSpan: %v", err)
		}
	})

	t.Run("Unit: test ListTechnoQueryHandler \"experience\" success", func(t *testing.T) {
		listTechnoQuery := query2.ListTechnoQuery{
			Type: "experience",
		}

		handler := query2.NewListTechnoQueryHandler(&TechnoRepositoryTest{})

		_, err := handler.Handle(context.Background(), listTechnoQuery)
		if err != nil {
			t.Errorf("ErrorSpan: %v", err)
		}
	})

	t.Run("Unit: test ListTechnoQueryHandler \"formation\" success", func(t *testing.T) {
		listTechnoQuery := query2.ListTechnoQuery{
			Type: "formation",
		}

		handler := query2.NewListTechnoQueryHandler(&TechnoRepositoryTest{})

		_, err := handler.Handle(context.Background(), listTechnoQuery)
		if err != nil {
			t.Errorf("ErrorSpan: %v", err)
		}
	})

	t.Run("Unit: test ListTechnoQueryHandler \"lang\" success", func(t *testing.T) {
		listTechnoQuery := query2.ListTechnoQuery{
			Type: "lang",
		}

		handler := query2.NewListTechnoQueryHandler(&TechnoRepositoryTest{})

		_, err := handler.Handle(context.Background(), listTechnoQuery)
		if err != nil {
			t.Errorf("ErrorSpan: %v", err)
		}
	})

	t.Run("Unit: test ListTechnoQueryHandler \"other\" success", func(t *testing.T) {
		listTechnoQuery := query2.ListTechnoQuery{
			Type: "other",
		}

		handler := query2.NewListTechnoQueryHandler(&TechnoRepositoryTest{})

		_, err := handler.Handle(context.Background(), listTechnoQuery)
		if err != nil {
			t.Errorf("ErrorSpan: %v", err)
		}
	})

	t.Run("Unit: test ListTechnoQueryHandler non existent type success", func(t *testing.T) {
		listTechnoQuery := query2.ListTechnoQuery{
			Type: "fake",
		}

		handler := query2.NewListTechnoQueryHandler(&TechnoRepositoryTest{})

		_, err := handler.Handle(context.Background(), listTechnoQuery)
		if err != nil {
			t.Errorf("ErrorSpan: %v", err)
		}
	})

	t.Run("Unit: test ListTechnoQueryHandler \"stack\" failed", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
		listTechnoQuery := query2.ListTechnoQuery{
			Type: "stack",
		}

		handler := query2.NewListTechnoQueryHandler(mocked.TechnoRepository)

		_, err := handler.Handle(context.Background(), listTechnoQuery)

		assert.Error(t, err, "error fetching stack: error")
	})
}

type TechnoRepositoryTest struct{}

func (m *TechnoRepositoryTest) FetchStack(ctx context.Context) (map[string]any, error) {
	_ = ctx
	return map[string]any{}, nil
}

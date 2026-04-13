package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	"go.opentelemetry.io/otel/trace/noop"
	"gotest.tools/assert"
)

func init() {
	l, err := logger.NewLogger(logger.Config{Level: "debug"})
	if err != nil {
		panic(err)
	}
	_, _ = observability.NewTelemetry(context.Background(), observability.Config{}, l)
}

// failCloser simulates an io.Closer that returns an error on Close.
type failCloser struct{}

func (f *failCloser) Close() error {
	return errors.New("close error")
}

// successCloser simulates an io.Closer that closes without error.
type successCloser struct{}

func (s *successCloser) Close() error {
	return nil
}

func TestCloseStmt(t *testing.T) {
	t.Run("Unit: test closeStmt with nil stmt does not panic", func(t *testing.T) {
		span := noop.Span{}

		closeStmt(context.Background(), nil, span)
	})

	t.Run("Unit: test closeStmt with valid closer succeeds", func(t *testing.T) {
		span := noop.Span{}

		closeStmt(context.Background(), &successCloser{}, span)
	})

	t.Run("Unit: test closeStmt with failing closer logs error", func(t *testing.T) {
		span := noop.Span{}

		closeStmt(context.Background(), &failCloser{}, span)
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("Unit: test errorResponse wraps error with message", func(t *testing.T) {
		span := noop.Span{}
		err := errorResponse("test error", span, sql.ErrNoRows)

		assert.ErrorContains(t, err, "test error")
		assert.ErrorContains(t, err, "no rows")
	})
}

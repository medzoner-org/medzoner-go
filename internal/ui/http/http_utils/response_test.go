package http_utils_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/ui/http/http_utils"
	"gotest.tools/assert"
)

func init() {
	l, err := logger.NewLogger(logger.Config{Level: "debug"})
	if err != nil {
		panic(err)
	}
	_, _ = observability.NewTelemetry(context.Background(), observability.Config{}, l)
}

func TestResponseError(t *testing.T) {
	t.Run("Unit: test ResponseError sets status 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, span := observability.StartSpan(context.Background(), "test")
		defer span.End()

		http_utils.ResponseError(w, errors.New("something went wrong"), http.StatusInternalServerError, span)

		assert.Equal(t, w.Code, http.StatusInternalServerError)
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
		assert.Equal(t, w.Header().Get("X-Content-Type-Options"), "nosniff")
		assert.Equal(t, w.Body.String(), "something went wrong")
	})

	t.Run("Unit: test ResponseError sets status 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, span := observability.StartSpan(context.Background(), "test")
		defer span.End()

		http_utils.ResponseError(w, errors.New("bad request"), http.StatusBadRequest, span)

		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Equal(t, w.Body.String(), "bad request")
	})

	t.Run("Unit: test ResponseError sets status 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, span := observability.StartSpan(context.Background(), "test")
		defer span.End()

		http_utils.ResponseError(w, errors.New("not found"), http.StatusNotFound, span)

		assert.Equal(t, w.Code, http.StatusNotFound)
		assert.Equal(t, w.Body.String(), "not found")
	})
}


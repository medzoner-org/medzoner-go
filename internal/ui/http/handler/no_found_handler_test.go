package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gohttp "github.com/Medzoner/gomedz/pkg/http"
	"github.com/Medzoner/medzoner-go/internal/ui/http/handler"
	"gotest.tools/assert"
)

// fakeNotFoundRenderer implements gohttp.Renderer for NotFoundHandler tests
type fakeNotFoundRenderer struct {
	err error
}

func (f *fakeNotFoundRenderer) Render(_ io.Writer, _ string, _ any, _ context.Context) error {
	return f.err
}

func TestNotFoundHandler_Handle(t *testing.T) {
	t.Run("Unit: test NotFoundHandler returns 404", func(t *testing.T) {
		renderer := &fakeNotFoundRenderer{}
		h := handler.NewNotFoundHandler(renderer)

		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		w := httptest.NewRecorder()

		h.Handle(w, req)

		assert.Equal(t, w.Code, http.StatusNotFound)
	})

	t.Run("Unit: test NotFoundHandler render error returns 500", func(t *testing.T) {
		renderer := &fakeNotFoundRenderer{err: errors.New("template error")}
		h := handler.NewNotFoundHandler(renderer)

		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		w := httptest.NewRecorder()

		h.Handle(w, req)

		assert.Equal(t, w.Code, http.StatusNotFound)
	})

	t.Run("Unit: test NotFoundHandler sets TOR-HOST header in view", func(t *testing.T) {
		renderer := &fakeNotFoundRenderer{}
		h := handler.NewNotFoundHandler(renderer)

		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		req.Header.Set("TOR-HOST", "test-tor-host")
		w := httptest.NewRecorder()

		h.Handle(w, req)

		assert.Equal(t, w.Code, http.StatusNotFound)
	})
}

// Ensure fakeNotFoundRenderer implements gohttp.Renderer
var _ gohttp.Renderer = (*fakeNotFoundRenderer)(nil)

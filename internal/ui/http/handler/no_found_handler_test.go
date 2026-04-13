package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Medzoner/medzoner-go/internal/ui/http/handler"
	mocks "github.com/Medzoner/medzoner-go/test"
	"go.uber.org/mock/gomock"
	"gotest.tools/assert"
)

func TestNotFoundHandler_Handle(t *testing.T) {
	t.Run("Unit: test NotFoundHandler returns 404", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Templater.EXPECT().Render("404", gomock.Any(), gomock.Any()).Return(nil).Times(1)

		h := handler.NewNotFoundHandler(mocked.Templater)

		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		w := httptest.NewRecorder()

		h.Handle(w, req)

		assert.Equal(t, w.Code, http.StatusNotFound)
	})

	t.Run("Unit: test NotFoundHandler render error returns 500", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Templater.EXPECT().Render("404", gomock.Any(), gomock.Any()).Return(errors.New("template error")).Times(1)

		h := handler.NewNotFoundHandler(mocked.Templater)

		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		w := httptest.NewRecorder()

		h.Handle(w, req)

		assert.Equal(t, w.Code, http.StatusNotFound)
	})

	t.Run("Unit: test NotFoundHandler sets TOR-HOST header in view", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Templater.EXPECT().Render("404", gomock.Any(), gomock.Any()).Return(nil).Times(1)

		h := handler.NewNotFoundHandler(mocked.Templater)

		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		req.Header.Set("TOR-HOST", "test-tor-host")
		w := httptest.NewRecorder()

		h.Handle(w, req)

		assert.Equal(t, w.Code, http.StatusNotFound)
	})
}

package handler

import (
	"fmt"
	"net/http"

	gohttp "github.com/Medzoner/gomedz/pkg/http"
	"github.com/Medzoner/gomedz/pkg/observability"
)

// NotFoundView NotFoundView
type NotFoundView struct {
	Locale          string
	PageTitle       string
	TorHost         string
	PageDescription string
}

// NotFoundHandler NotFoundHandler
type NotFoundHandler struct {
	Renderer gohttp.Renderer
}

// NewNotFoundHandler NewNotFoundHandler
func NewNotFoundHandler(renderer gohttp.Renderer) *NotFoundHandler {
	return &NotFoundHandler{
		Renderer: renderer,
	}
}

// Handle handles NotFoundHandler
func (h *NotFoundHandler) Handle(w http.ResponseWriter, r *http.Request) {
	_, span := observability.StartSpan(r.Context(), "NotFoundHandler.Handle")
	defer span.End()

	view := &NotFoundView{
		Locale:          "fr",
		PageTitle:       "MedZoner.com - Not Found",
		TorHost:         r.Header.Get("TOR-HOST"),
		PageDescription: "MedZoner.com - Not Found",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	if err := h.Renderer.Render(w, "404", view, r.Context()); err != nil {
		span.RecordError(fmt.Errorf("error rendering 404 template: %w", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

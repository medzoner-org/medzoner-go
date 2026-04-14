package handler

import (
	"net/http"

	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/ui/http/http_utils"
	"github.com/Medzoner/medzoner-go/internal/ui/http/templater"
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
	Template templater.Templater
}

// NewNotFoundHandler NewNotFoundHandler
func NewNotFoundHandler(template templater.Templater) *NotFoundHandler {
	return &NotFoundHandler{
		Template: template,
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

	w.WriteHeader(http.StatusNotFound)

	if err := h.Template.Render("404", view, w); err != nil {
		http_utils.ResponseError(w, err, http.StatusInternalServerError, span)
		return
	}
}

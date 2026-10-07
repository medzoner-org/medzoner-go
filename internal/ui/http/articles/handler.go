package articles

import (
	"fmt"
	"net/http"

	gohttp "github.com/medzoner-org/gomedz/pkg/http"
	"github.com/medzoner-org/gomedz/pkg/observability"
)

const (
	Path  = "/articles/after-wire-whyor"
	Title = "After Wire: compile-time dependency injection with whyor"
)

type View struct {
	Locale          string
	PageTitle       string
	PageDescription string
	TorHost         string
	CanonicalURL    string
	SocialImage     string
}

// Handler serves the editorial article without database or contact-form dependencies.
type Handler struct{}

func NewHandler() Handler      { return Handler{} }
func (Handler) Prefix() string { return "/articles" }

func (h Handler) Register(r gohttp.Router[any]) {
	r.Get("/after-wire-whyor", h.Whyor, gohttp.Options{})
}

func (Handler) Whyor(c *gohttp.Context, _ struct{}) error {
	_, span := observability.StartSpan(c.Context(), "Articles.Whyor")
	defer span.End()
	view := View{
		Locale:          "en",
		PageTitle:       Title + " | Medzoner",
		PageDescription: "Why I built whyor, a compile-time dependency injection tool for Go: plain generated code, a Wire migration example, and honest limits.",
		CanonicalURL:    "https://www.medzoner.com" + Path,
		SocialImage:     "https://www.medzoner.com/public/images/whyor-social.png",
	}
	if err := c.HTML(http.StatusOK, "article_whyor", view); err != nil {
		span.RecordError(err)
		return fmt.Errorf("render whyor article: %w", err)
	}
	return nil
}

package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Medzoner/gomedz/pkg/captcha"
	http2 "github.com/Medzoner/gomedz/pkg/http"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/gomedz/pkg/validation"
	command2 "github.com/Medzoner/medzoner-go/internal/application/command"
	"github.com/Medzoner/medzoner-go/internal/ui/http/http_utils"
)

// IndexView IndexView
type IndexView struct {
	Locale           string
	PageTitle        string
	TorHost          string
	Errors           any
	RecaptchaSiteKey string
	PageDescription  string
	FormMessage      string
}

// IndexHandler IndexHandler
type IndexHandler struct {
	CreateContactCommandHandler command2.CreateContactCommandHandler
	Validation                  validation.Validater
	Recaptcha                   captcha.Captcher
}

// NewIndexHandler NewIndexHandler
func NewIndexHandler(
	createContactCommandHandler command2.CreateContactCommandHandler,
	validation validation.Validater,
	recaptcha captcha.Captcher,
) IndexHandler {
	return IndexHandler{
		CreateContactCommandHandler: createContactCommandHandler,
		Validation:                  validation,
		Recaptcha:                   recaptcha,
	}
}

func (h IndexHandler) Prefix() string {
	return "/"
}

func (h IndexHandler) Register(r http2.Router[any]) {
	r.Get("/", h.Index, http2.Options{})
	r.Post("/", h.Index, http2.Options{})

	r.Get("/robots.txt", h.serveStaticFile("./public/robots.txt"), http2.Options{})
	r.Get("/sitemap.xml", h.serveStaticFile("./public/sitemap.xml"), http2.Options{})

	r.StaticFS("/public", http.Dir("./public"), http2.Options{})
}

func (h IndexHandler) serveStaticFile(filePath string) func(c *http2.Context, _ struct{}) error {
	return func(c *http2.Context, _ struct{}) error {
		http.ServeFile(c.Writer(), c.Request(), filePath)
		return nil
	}
}

func (h IndexHandler) processRequest(request *http.Request) (err error) {
	recaptchaResponse, responseFound := request.Form["g-captcha-response"]
	if responseFound {
		result, err := h.Recaptcha.Confirm(request.RemoteAddr, recaptchaResponse[0])
		if err != nil {
			return fmt.Errorf("captcha server error: %w", err)
		}
		if !result {
			return fmt.Errorf("captcha was incorrect; try again")
		}
	}
	return nil
}

// Index Index
func (h IndexHandler) Index(c *http2.Context, _ struct{}) error {
	w := c.Writer()
	r := c.Request()

	ctx, span := observability.StartSpan(c.Context(), "IndexHandler.IndexHandle")
	defer span.End()

	view, err := h.initView(ctx, r)
	if err != nil {
		http_utils.ResponseError(w, err, http.StatusInternalServerError, span)
		return nil
	}
	statusCode := http.StatusOK
	if r.Method == "POST" && r.FormValue("submit") == "" {
		if err = h.processRequest(r); err != nil {
			http.Redirect(w, r, "/#contact?msg=\"Recaptcha was incorrect; try again.\"", http.StatusSeeOther)
			return nil
		}
		createContactCommand := command2.CreateContactCommand{
			DateAdd: time.Now(),
			Name:    r.FormValue("name"),
			Email:   r.FormValue("email"),
			Message: r.FormValue("message"),
		}

		validationError := h.Validation.Struct(createContactCommand)
		if validationError == nil {
			if err = h.CreateContactCommandHandler.Handle(ctx, createContactCommand); err != nil {
				return fmt.Errorf("error during create contact command handling: %w", err)
			}
			http.Redirect(w, r, "/#contact", http.StatusSeeOther)
			return nil
		}
		statusCode = http.StatusBadRequest
	}

	if err = c.HTML(statusCode, "index", view); err != nil {
		return fmt.Errorf("error during render template: %w", err)
	}

	return nil
}

func (h IndexHandler) initView(ctx context.Context, request *http.Request) (IndexView, error) {
	return IndexView{
		Locale:           "fr",
		PageTitle:        "MedZoner.com",
		TorHost:          request.Header.Get("TOR-HOST"),
		RecaptchaSiteKey: h.Recaptcha.GetSiteKey(),
		PageDescription:  "Mehdi YOUB - Développeur Web Full Stack - NestJS Symfony Golang VueJS",
		FormMessage:      "",
	}, nil
}

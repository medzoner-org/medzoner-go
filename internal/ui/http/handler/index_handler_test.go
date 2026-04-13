package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	gohttp "github.com/Medzoner/gomedz/pkg/http"
	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	command2 "github.com/Medzoner/medzoner-go/internal/application/command"
	"github.com/Medzoner/medzoner-go/internal/application/event"
	query2 "github.com/Medzoner/medzoner-go/internal/application/query"
	"github.com/Medzoner/medzoner-go/internal/ui/http/handler"
	mocks "github.com/Medzoner/medzoner-go/test"
	"go.uber.org/mock/gomock"
	"gotest.tools/assert"
)

func init() {
	l, err := logger.NewLogger(logger.Config{Level: "debug"})
	if err != nil {
		panic(err)
	}
	_, _ = observability.NewTelemetry(context.Background(), observability.Config{}, l)
}

func TestIndexHandler_Index_GET(t *testing.T) {
	t.Run("Unit: test IndexHandler GET success", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{"Go": "ok"}, nil).Times(1)
		mocked.Templater.EXPECT().Render("index", gomock.Any(), gomock.Any()).Return(nil).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusOK)
	})
}

func TestIndexHandler_Index_GET_FetchStackError(t *testing.T) {
	t.Run("Unit: test IndexHandler GET error fetch stack", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(nil, errors.New("db error")).Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusInternalServerError)
	})
}

func TestIndexHandler_Index_GET_RenderError(t *testing.T) {
	t.Run("Unit: test IndexHandler GET error render template", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{}, nil).Times(1)
		mocked.Templater.EXPECT().Render("index", gomock.Any(), gomock.Any()).Return(errors.New("render error")).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.ErrorContains(t, err, "error during render template")
	})
}

func TestIndexHandler_Index_POST_ValidationError(t *testing.T) {
	t.Run("Unit: test IndexHandler POST with validation error", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{}, nil).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Validater.EXPECT().Struct(gomock.Any()).Return(errors.New("validation error")).Times(1)
		mocked.Templater.EXPECT().Render("index", gomock.Any(), gomock.Any()).Return(nil).Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		form := url.Values{}
		form.Set("name", "test")
		form.Set("email", "bad")
		form.Set("message", "hello")
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.NilError(t, err)
	})
}

func TestIndexHandler_Index_POST_Success(t *testing.T) {
	t.Run("Unit: test IndexHandler POST success", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{}, nil).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Validater.EXPECT().Struct(gomock.Any()).Return(nil).Times(1)
		mocked.ContactRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		mocked.Mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(true, nil).Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		form := url.Values{}
		form.Set("name", "test")
		form.Set("email", "test@example.com")
		form.Set("message", "hello world")
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusSeeOther)
	})
}

func TestIndexHandler_Index_POST_CaptchaError(t *testing.T) {
	t.Run("Unit: test IndexHandler POST captcha error", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{}, nil).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Captcher.EXPECT().Confirm(gomock.Any(), gomock.Any()).Return(false, errors.New("captcha error")).Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		form := url.Values{}
		form.Set("name", "test")
		form.Set("email", "test@example.com")
		form.Set("message", "hello")
		form.Set("g-captcha-response", "bad-token")
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusSeeOther)
	})
}

func TestIndexHandler_Index_POST_CommandHandlerError(t *testing.T) {
	t.Run("Unit: test IndexHandler POST command handler error", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{}, nil).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Validater.EXPECT().Struct(gomock.Any()).Return(nil).Times(1)
		mocked.ContactRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("save error")).Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		form := url.Values{}
		form.Set("name", "test")
		form.Set("email", "test@example.com")
		form.Set("message", "hello")
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.ErrorContains(t, err, "error during create contact command handling")
	})
}

func TestIndexHandler_Prefix(t *testing.T) {
	t.Run("Unit: test IndexHandler Prefix returns /", func(t *testing.T) {
		h := handler.NewIndexHandler(nil, query2.ListTechnoQueryHandler{}, command2.CreateContactCommandHandler{}, nil, nil)
		assert.Equal(t, h.Prefix(), "/")
	})
}

func TestIndexHandler_Index_POST_CaptchaConfirmFailed(t *testing.T) {
	t.Run("Unit: test IndexHandler POST captcha confirm returns false", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{}, nil).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Captcher.EXPECT().Confirm(gomock.Any(), gomock.Any()).Return(false, nil).Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		form := url.Values{}
		form.Set("name", "test")
		form.Set("email", "test@example.com")
		form.Set("message", "hello")
		form.Set("g-captcha-response", "bad-token")
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusSeeOther)
	})
}

func TestIndexHandler_Index_POST_WithSubmit(t *testing.T) {
	t.Run("Unit: test IndexHandler POST with submit skips processing", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]any{}, nil).Times(1)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Templater.EXPECT().Render("index", gomock.Any(), gomock.Any()).Return(nil).Times(1)

		h := handler.NewIndexHandler(
			mocked.Templater,
			query2.NewListTechnoQueryHandler(mocked.TechnoRepository),
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		form := url.Values{}
		form.Set("name", "test")
		form.Set("submit", "Send")
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)

		err := h.Index(ctx)

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusOK)
	})
}

func TestIndexHandler_Register(t *testing.T) {
	t.Run("Unit: test IndexHandler Register registers routes", func(t *testing.T) {
		h := handler.NewIndexHandler(nil, query2.ListTechnoQueryHandler{}, command2.CreateContactCommandHandler{}, nil, nil)
		r := &fakeRouter{}

		h.Register(r)

		assert.Equal(t, r.getCalled, true)
		assert.Equal(t, r.postCalled, true)
		assert.Equal(t, r.staticFSCalled, true)
	})
}

// fakeRouter implements gohttp.Router for testing Register
type fakeRouter struct {
	getCalled      bool
	postCalled     bool
	staticFSCalled bool
}

func (f *fakeRouter) Use(_ ...gohttp.Middleware)                               {}
func (f *fakeRouter) Static(_, _ string)                                       {}
func (f *fakeRouter) StaticFS(_ string, _ http.FileSystem, _ gohttp.Options)   { f.staticFSCalled = true }
func (f *fakeRouter) Get(_ string, _ gohttp.HandlerFunc, _ gohttp.Options)     { f.getCalled = true }
func (f *fakeRouter) Post(_ string, _ gohttp.HandlerFunc, _ gohttp.Options)    { f.postCalled = true }
func (f *fakeRouter) Put(_ string, _ gohttp.HandlerFunc, _ gohttp.Options)     {}
func (f *fakeRouter) Delete(_ string, _ gohttp.HandlerFunc, _ gohttp.Options)  {}
func (f *fakeRouter) Patch(_ string, _ gohttp.HandlerFunc, _ gohttp.Options)   {}
func (f *fakeRouter) Any(_ string, _ gohttp.HandlerFunc, _ gohttp.Options)     {}
func (f *fakeRouter) Options(_ string, _ gohttp.HandlerFunc, _ gohttp.Options) {}
func (f *fakeRouter) Group(_ string, _ ...gohttp.Middleware) gohttp.Router     { return f }

package handler_test

import (
	"context"
	"errors"
	"io"
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

// fakeRenderer implements gohttp.Renderer for tests
type fakeRenderer struct {
	err error
}

func (f *fakeRenderer) Render(_ io.Writer, _ string, _ any, _ context.Context) error {
	return f.err
}

func newTestContext(w *httptest.ResponseRecorder, req *http.Request, renderer gohttp.Renderer) *gohttp.Context {
	ctx := gohttp.NewContext(w, req)
	ctx.SetRenderer(renderer)
	return ctx
}

func TestIndexHandler_Index_GET(t *testing.T) {
	t.Run("Unit: test IndexHandler GET success", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)

		h := handler.NewIndexHandler(
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := newTestContext(w, req, &fakeRenderer{})

		err := h.Index(ctx, struct{}{})

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusOK)
	})
}

func TestIndexHandler_Index_GET_RenderError(t *testing.T) {
	t.Run("Unit: test IndexHandler GET error render template", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)

		h := handler.NewIndexHandler(
			command2.NewCreateContactCommandHandler(mocked.ContactRepository, event.ContactCreatedEventHandler{Mailer: mocked.Mailer}),
			mocked.Validater,
			mocked.Captcher,
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		ctx := newTestContext(w, req, &fakeRenderer{err: errors.New("render error")})

		err := h.Index(ctx, struct{}{})

		assert.ErrorContains(t, err, "error during render template")
	})
}

func TestIndexHandler_Index_POST_ValidationError(t *testing.T) {
	t.Run("Unit: test IndexHandler POST with validation error", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Validater.EXPECT().Struct(gomock.Any()).Return(errors.New("validation error")).Times(1)

		h := handler.NewIndexHandler(
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
		ctx := newTestContext(w, req, &fakeRenderer{})

		err := h.Index(ctx, struct{}{})

		assert.NilError(t, err)
	})
}

func TestIndexHandler_Index_POST_Success(t *testing.T) {
	t.Run("Unit: test IndexHandler POST success", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Validater.EXPECT().Struct(gomock.Any()).Return(nil).Times(1)
		mocked.ContactRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		mocked.Mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(true, nil).Times(1)

		h := handler.NewIndexHandler(
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
		ctx := newTestContext(w, req, &fakeRenderer{})

		err := h.Index(ctx, struct{}{})

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusSeeOther)
	})
}

func TestIndexHandler_Index_POST_CaptchaError(t *testing.T) {
	t.Run("Unit: test IndexHandler POST captcha error", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Captcher.EXPECT().Confirm(gomock.Any(), gomock.Any()).Return(false, errors.New("captcha error")).Times(1)

		h := handler.NewIndexHandler(
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
		ctx := newTestContext(w, req, &fakeRenderer{})

		err := h.Index(ctx, struct{}{})

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusSeeOther)
	})
}

func TestIndexHandler_Index_POST_CommandHandlerError(t *testing.T) {
	t.Run("Unit: test IndexHandler POST command handler error", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Validater.EXPECT().Struct(gomock.Any()).Return(nil).Times(1)
		mocked.ContactRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("save error")).Times(1)

		h := handler.NewIndexHandler(
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
		ctx := newTestContext(w, req, &fakeRenderer{})

		err := h.Index(ctx, struct{}{})

		assert.ErrorContains(t, err, "error during create contact command handling")
	})
}

func TestIndexHandler_Prefix(t *testing.T) {
	t.Run("Unit: test IndexHandler Prefix returns /", func(t *testing.T) {
		h := handler.NewIndexHandler(command2.CreateContactCommandHandler{}, nil, nil)
		assert.Equal(t, h.Prefix(), "/")
	})
}

func TestIndexHandler_Index_POST_CaptchaConfirmFailed(t *testing.T) {
	t.Run("Unit: test IndexHandler POST captcha confirm returns false", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)
		mocked.Captcher.EXPECT().Confirm(gomock.Any(), gomock.Any()).Return(false, nil).Times(1)

		h := handler.NewIndexHandler(
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
		ctx := newTestContext(w, req, &fakeRenderer{})

		err := h.Index(ctx, struct{}{})

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusSeeOther)
	})
}

func TestIndexHandler_Index_POST_WithSubmit(t *testing.T) {
	t.Run("Unit: test IndexHandler POST with submit skips processing", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key").Times(1)

		h := handler.NewIndexHandler(
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
		ctx := newTestContext(w, req, &fakeRenderer{})

		err := h.Index(ctx, struct{}{})

		assert.NilError(t, err)
		assert.Equal(t, w.Code, http.StatusOK)
	})
}

func TestIndexHandler_Register(t *testing.T) {
	t.Run("Unit: test IndexHandler Register registers routes", func(t *testing.T) {
		h := handler.NewIndexHandler(command2.CreateContactCommandHandler{}, nil, nil)
		r := &fakeRouter{}

		h.Register(r)

		assert.Equal(t, r.getCalled, true)
		assert.Equal(t, r.postCalled, true)
		assert.Equal(t, r.staticFSCalled, true)
	})
}

// fakeRouter implements gohttp.Router[any] for testing Register
type fakeRouter struct {
	getCalled      bool
	postCalled     bool
	staticFSCalled bool
}

func (f *fakeRouter) Use(_ ...gohttp.Middleware[struct{}])                   {}
func (f *fakeRouter) UseBody(_ ...gohttp.Middleware[any])                    {}
func (f *fakeRouter) BodyMiddlewares() []gohttp.Middleware[any]              { return nil }
func (f *fakeRouter) Static(_, _ string)                                     {}
func (f *fakeRouter) StaticFS(_ string, _ http.FileSystem, _ gohttp.Options) { f.staticFSCalled = true }
func (f *fakeRouter) Get(_ string, _ gohttp.HandlerFunc[struct{}], _ gohttp.Options) {
	f.getCalled = true
}
func (f *fakeRouter) Post(_ string, _ gohttp.HandlerFunc[struct{}], _ gohttp.Options) {
	f.postCalled = true
}
func (f *fakeRouter) Put(_ string, _ gohttp.HandlerFunc[struct{}], _ gohttp.Options)      {}
func (f *fakeRouter) Delete(_ string, _ gohttp.HandlerFunc[struct{}], _ gohttp.Options)   {}
func (f *fakeRouter) Patch(_ string, _ gohttp.HandlerFunc[struct{}], _ gohttp.Options)    {}
func (f *fakeRouter) Any(_ string, _ gohttp.HandlerFunc[struct{}], _ gohttp.Options)      {}
func (f *fakeRouter) Options(_ string, _ gohttp.HandlerFunc[struct{}], _ gohttp.Options)  {}
func (f *fakeRouter) Group(_ string, _ ...gohttp.Middleware[struct{}]) gohttp.Router[any] { return f }

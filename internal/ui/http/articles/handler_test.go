package articles_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	markdown "github.com/Medzoner/medzoner-go/internal/adapters/editorial"
	"github.com/Medzoner/medzoner-go/internal/application/query"
	"github.com/Medzoner/medzoner-go/internal/resources"
	"github.com/Medzoner/medzoner-go/internal/ui/http/articles"
	gohttp "github.com/medzoner-org/gomedz/pkg/http"
	"gotest.tools/assert"
)

const (
	whyorPath  = "/articles/after-wire-whyor"
	whyorTitle = "After Wire: compile-time dependency injection with whyor"
	esp32Path  = "/articles/resilient-multi-client-esp32-camera"
	esp32Title = "Building a resilient multi-client ESP32 camera"
)

func newHandler(t *testing.T) articles.Handler {
	t.Helper()
	files, err := resources.EditorialFiles()
	assert.NilError(t, err)
	repo, err := markdown.NewRepository(context.Background(), files)
	assert.NilError(t, err)
	queries := query.NewEditorial(repo, func() time.Time { return time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC) })
	return articles.NewHandler(queries, markdown.NewRenderer())
}

type failingRenderer struct{ err error }

func (r failingRenderer) Render(io.Writer, string, any, context.Context) error { return r.err }

func TestWhyorArticle(t *testing.T) {
	t.Run("Unit: test article content and metadata", func(t *testing.T) {
		renderer := gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl"))
		req := httptest.NewRequest(http.MethodGet, whyorPath+"?title=UNTRUSTED_TITLE", nil)
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req, gohttp.Param{Key: "slug", Value: "after-wire-whyor"})
		ctx.SetRenderer(renderer)
		assert.NilError(t, newHandler(t).Show(ctx, struct{}{}))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Assert(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/html"))
		body := w.Body.String()
		for _, want := range []string{
			`<html lang="en">`, whyorTitle,
			`rel="canonical" href="https://www.medzoner.com/articles/after-wire-whyor"`,
			`property="og:type" content="article"`, `name="twitter:card" content="summary_large_image"`,
			`/public/images/whyor-social.png`, `/public/images/whyor-flow.svg`,
			`whyor.Bind[Store, *MemoryStore]()`, `pre-1.0`, `https://github.com/Medzoner/whyor`,
			`class="article-note"`, `class="article-figure"`, `<figcaption>`, `class="article-cta"`, `href="#why"`,
		} {
			assert.Assert(t, strings.Contains(body, want), "missing content: %s", want)
		}
		assert.Assert(t, !strings.Contains(body, "UNTRUSTED_TITLE"))
		data := regexp.MustCompile(`(?s)<script type="application/ld\+json">\s*(.*?)\s*</script>`).FindStringSubmatch(body)
		assert.Equal(t, len(data), 2)
		var schema map[string]any
		assert.NilError(t, json.Unmarshal([]byte(data[1]), &schema))
		assert.Equal(t, schema["@type"], "TechArticle")
		assert.Equal(t, schema["inLanguage"], "en")
	})
	t.Run("Unit: test article render error is wrapped", func(t *testing.T) {
		cause := errors.New("template unavailable")
		ctx := gohttp.NewContext(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, whyorPath, nil), gohttp.Param{Key: "slug", Value: "after-wire-whyor"})
		ctx.SetRenderer(failingRenderer{err: cause})
		err := newHandler(t).Show(ctx, struct{}{})
		assert.ErrorContains(t, err, "render article")
		assert.Assert(t, errors.Is(err, cause))
	})
}

func TestESP32CamArticle(t *testing.T) {
	t.Run("Unit: test ESP32-CAM article content and metadata", func(t *testing.T) {
		renderer := gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl"))
		req := httptest.NewRequest(http.MethodGet, esp32Path+"?title=UNTRUSTED_TITLE", nil)
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req, gohttp.Param{Key: "slug", Value: "resilient-multi-client-esp32-camera"})
		ctx.SetRenderer(renderer)
		assert.NilError(t, newHandler(t).Show(ctx, struct{}{}))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Assert(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/html"))
		body := w.Body.String()
		for _, want := range []string{
			`<html lang="en">`, esp32Title,
			`rel="canonical" href="https://www.medzoner.com/articles/resilient-multi-client-esp32-camera"`,
			`property="og:type" content="article"`, `name="twitter:card" content="summary_large_image"`,
			`/public/images/esp32cam-social.png`, `/public/images/esp32cam-flow.svg`,
			`/public/images/esp32cam-dashboard.png`, `http://camera.local:81/stream`,
			`/public/images/esp32cam-clients.svg`,
			`https://github.com/Medzoner/domoticz-esp32cam`, `no TLS`,
			`class="article-note"`, `class="article-figure"`, `<figcaption>`, `class="article-cta"`, `href="#shared-stream"`,
		} {
			assert.Assert(t, strings.Contains(body, want), "missing content: %s", want)
		}
		assert.Assert(t, !strings.Contains(body, "UNTRUSTED_TITLE"))
		data := regexp.MustCompile(`(?s)<script type="application/ld\+json">\s*(.*?)\s*</script>`).FindStringSubmatch(body)
		assert.Equal(t, len(data), 2)
		var schema map[string]any
		assert.NilError(t, json.Unmarshal([]byte(data[1]), &schema))
		assert.Equal(t, schema["@type"], "TechArticle")
		assert.Equal(t, schema["inLanguage"], "en")
	})
	t.Run("Unit: test ESP32-CAM article render error is wrapped", func(t *testing.T) {
		cause := errors.New("template unavailable")
		ctx := gohttp.NewContext(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, esp32Path, nil), gohttp.Param{Key: "slug", Value: "resilient-multi-client-esp32-camera"})
		ctx.SetRenderer(failingRenderer{err: cause})
		err := newHandler(t).Show(ctx, struct{}{})
		assert.ErrorContains(t, err, "render article")
		assert.Assert(t, errors.Is(err, cause))
	})
}

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

	"github.com/Medzoner/medzoner-go/internal/ui/http/articles"
	gohttp "github.com/medzoner-org/gomedz/pkg/http"
	"gotest.tools/assert"
)

type failingRenderer struct{ err error }

func (r failingRenderer) Render(io.Writer, string, any, context.Context) error { return r.err }

func TestWhyorArticle(t *testing.T) {
	t.Run("Unit: test article content and metadata", func(t *testing.T) {
		renderer := gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl"))
		req := httptest.NewRequest(http.MethodGet, articles.Path+"?title=UNTRUSTED_TITLE", nil)
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, req)
		ctx.SetRenderer(renderer)
		assert.NilError(t, articles.NewHandler().Whyor(ctx, struct{}{}))
		assert.Equal(t, w.Code, http.StatusOK)
		assert.Assert(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/html"))
		body := w.Body.String()
		for _, want := range []string{
			`<html lang="en">`, articles.Title,
			`rel="canonical" href="https://www.medzoner.com/articles/after-wire-whyor"`,
			`property="og:type" content="article"`, `name="twitter:card" content="summary_large_image"`,
			`/public/images/whyor-social.png`, `/public/images/whyor-flow.svg`,
			`whyor.Bind[Store, *MemoryStore]()`, `pre-1.0`, `https://github.com/Medzoner/whyor`,
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
		ctx := gohttp.NewContext(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, articles.Path, nil))
		ctx.SetRenderer(failingRenderer{err: cause})
		err := articles.NewHandler().Whyor(ctx, struct{}{})
		assert.ErrorContains(t, err, "render whyor article")
		assert.Assert(t, errors.Is(err, cause))
	})
}

package index_test

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Medzoner/medzoner-go/internal/ui/http/index"
	mocks "github.com/Medzoner/medzoner-go/test"
	gohttp "github.com/medzoner-org/gomedz/pkg/http"
	"gotest.tools/assert"
)

func TestHomepageSEO(t *testing.T) {
	t.Run("Unit: test homepage metadata and semantic content", func(t *testing.T) {
		mocked := mocks.New(t)
		mocked.Captcher.EXPECT().GetSiteKey().Return("test-site-key")
		h := index.Handler{Recaptcha: mocked.Captcher}
		w := httptest.NewRecorder()
		renderer := gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl"))
		req := httptest.NewRequest(http.MethodGet, "/?title=UNTRUSTED_TITLE", nil)
		ctx := newTestContext(w, req, renderer)
		assert.NilError(t, h.Index(ctx, struct{}{}))
		assert.Equal(t, w.Code, http.StatusOK)
		body := w.Body.String()
		assert.Assert(t, strings.HasPrefix(strings.TrimSpace(body), "<!DOCTYPE html>"))
		for _, want := range []string{
			`<html class="no-js" lang="fr">`,
			`rel="canonical" href="https://www.medzoner.com/"`,
			`property="og:type" content="website"`,
			`name="twitter:card" content="summary"`,
			`<h1 class="title">Mehdi Youb</h1>`,
			`id="technos"`, `id="contact"`, `<main`, `</main>`,
			`hreflang="en"`, `Développeur Go`,
		} {
			assert.Assert(t, strings.Contains(body, want), "missing SEO content: %s", want)
		}
		assert.Equal(t, strings.Count(body, "<h1 "), 1)
		assert.Assert(t, !strings.Contains(body, "UNTRUSTED_TITLE"))
		data := regexp.MustCompile(`(?s)<script type="application/ld\+json">\s*(.*?)\s*</script>`).FindStringSubmatch(body)
		assert.Equal(t, len(data), 2)
		var schema struct {
			Context string           `json:"@context"`
			Graph   []map[string]any `json:"@graph"`
		}
		assert.NilError(t, json.Unmarshal([]byte(data[1]), &schema))
		assert.Equal(t, schema.Context, "https://schema.org")
		assert.Equal(t, len(schema.Graph), 2)
		assert.Equal(t, schema.Graph[0]["@type"], "Person")
		assert.Equal(t, schema.Graph[1]["@type"], "WebSite")
	})
}

func TestCrawlDiscovery(t *testing.T) {
	t.Run("Unit: test sitemap only lists canonical pages and robots advertises it", func(t *testing.T) {
		root := filepath.Join("..", "..", "..", "..", "public")
		data, err := os.ReadFile(filepath.Join(root, "sitemap.xml"))
		assert.NilError(t, err)
		var sitemap struct {
			URLs []struct {
				Location string `xml:"loc"`
			} `xml:"url"`
		}
		assert.NilError(t, xml.Unmarshal(data, &sitemap))
		assert.Equal(t, len(sitemap.URLs), 3)
		for i, want := range []string{
			"https://www.medzoner.com/",
			"https://www.medzoner.com/articles/after-wire-whyor",
			"https://www.medzoner.com/articles/resilient-multi-client-esp32-camera",
		} {
			assert.Equal(t, sitemap.URLs[i].Location, want)
		}
		robots, err := os.ReadFile(filepath.Join(root, "robots.txt"))
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(string(robots), "Sitemap: https://www.medzoner.com/sitemap.xml"))
	})
}

func TestNotFoundSEO(t *testing.T) {
	t.Run("Unit: test not found template excludes indexing", func(t *testing.T) {
		renderer := gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl"))
		var body bytes.Buffer
		assert.NilError(t, renderer.Render(&body, "404", nil, context.Background()))
		assert.Assert(t, strings.HasPrefix(strings.TrimSpace(body.String()), "<!DOCTYPE html>"))
		assert.Assert(t, strings.Contains(body.String(), `name="robots" content="noindex"`))
		assert.Assert(t, !strings.Contains(body.String(), `rel="canonical"`))
	})
}

package articles_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	markdown "github.com/Medzoner/medzoner-go/internal/adapters/editorial"
	"github.com/Medzoner/medzoner-go/internal/application/query"
	"github.com/Medzoner/medzoner-go/internal/domains/editorial"
	"github.com/Medzoner/medzoner-go/internal/ui/http/articles"
	gohttp "github.com/medzoner-org/gomedz/pkg/http"
	"gotest.tools/assert"
)

type catalogue struct{ articles []editorial.Article }

func (r catalogue) All(context.Context) ([]editorial.Article, error) { return r.articles, nil }

func TestPublicCatalogue(t *testing.T) {
	now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	a := editorial.Article{Slug: "third-article", Title: "New & independent", Description: "A <description>", Status: "published", PublishedAt: now, Author: editorial.Author{Name: "Author", URL: "https://www.medzoner.com/"}, Category: "go", Tags: []string{"testing"}, Language: "en", Image: "/public/images/whyor-social.png", Markdown: "## New content {#stable}\n\nAdded without a handler."}
	draft, future := a, a
	draft.Slug, draft.Status = "private-draft", "draft"
	future.Slug, future.PublishedAt = "scheduled-secret", now.Add(time.Hour)
	h := articles.NewHandler(query.NewEditorial(catalogue{[]editorial.Article{a, draft, future}}, func() time.Time { return now }), markdown.NewRenderer())
	renderer := gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl"))

	for _, route := range []struct {
		path    string
		handler gohttp.HandlerFunc[struct{}]
		slug    string
		status  int
		want    string
	}{
		{"/articles", h.List, "", 200, "third-article"},
		{"/articles?category=go&tag=testing", h.List, "", 200, "third-article"},
		{"/articles?category=iot", h.List, "", 200, "Aucun article"},
		{"/articles?page=2", h.List, "", 200, "Aucun article"},
		{"/articles?page=0", h.List, "", 400, "Invalid page"},
		{"/articles?page=no", h.List, "", 400, "Invalid page"},
		{"/articles?page=9223372036854775807", h.List, "", 200, "Aucun article"},
		{"/articles/third-article", h.Show, "third-article", 200, "Added without a handler"},
		{"/articles/private-draft", h.Show, "private-draft", 404, `content="noindex"`},
		{"/articles/scheduled-secret", h.Show, "scheduled-secret", 404, `content="noindex"`},
		{"/articles/unknown", h.Show, "unknown", 404, `content="noindex"`},
		{"/sitemap.xml", h.Sitemap, "", 200, "third-article"},
		{"/feed.xml", h.Feed, "", 200, "third-article"},
	} {
		t.Run(fmt.Sprintf("Unit: test catalogue %s", route.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := gohttp.NewContext(w, httptest.NewRequest(http.MethodGet, route.path, nil), gohttp.Param{Key: "slug", Value: route.slug})
			ctx.SetRenderer(renderer)
			assert.NilError(t, route.handler(ctx, struct{}{}))
			assert.Equal(t, w.Code, route.status)
			assert.Assert(t, strings.Contains(w.Body.String(), route.want))
			assert.Assert(t, !strings.Contains(w.Body.String(), "private-draft"))
			assert.Assert(t, !strings.Contains(w.Body.String(), "scheduled-secret"))
			if route.path == "/articles/third-article" {
				assert.Assert(t, !strings.Contains(w.Body.String(), "Updated <time"))
			}
			if strings.HasSuffix(route.path, ".xml") {
				var doc any
				assert.NilError(t, xml.Unmarshal(w.Body.Bytes(), &doc))
				assert.Assert(t, strings.Contains(w.Header().Get("Content-Type"), "xml"))
			}
		})
	}
}

func TestMetadataEscaping(t *testing.T) {
	t.Run("Unit: test metadata cannot break out of JSON-LD or HTML", func(t *testing.T) {
		now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
		a := editorial.Article{Slug: "safe", Title: `</script><script>alert(1)</script>`, Description: `" onmouseover="alert(1)`, Language: "en", Status: "published", PublishedAt: now, Image: "/public/images/whyor-social.png", Markdown: "Safe content."}
		h := articles.NewHandler(query.NewEditorial(catalogue{[]editorial.Article{a}}, func() time.Time { return now }), markdown.NewRenderer())
		w := httptest.NewRecorder()
		ctx := gohttp.NewContext(w, httptest.NewRequest(http.MethodGet, "/articles/safe", nil), gohttp.Param{Key: "slug", Value: "safe"})
		ctx.SetRenderer(gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl")))
		assert.NilError(t, h.Show(ctx, struct{}{}))
		body := w.Body.String()
		assert.Assert(t, !strings.Contains(body, `<script>alert(1)</script>`))
		data := regexp.MustCompile(`(?s)<script type="application/ld\+json">\s*(.*?)\s*</script>`).FindStringSubmatch(body)
		assert.Equal(t, len(data), 2)
		var schema map[string]any
		assert.NilError(t, json.Unmarshal([]byte(data[1]), &schema))
		assert.Equal(t, schema["headline"], a.Title)
	})
}

func TestListingNavigation(t *testing.T) {
	t.Run("Unit: test pagination preserves category and tag filters", func(t *testing.T) {
		now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
		var all []editorial.Article
		for i := range 12 {
			all = append(all, editorial.Article{Slug: fmt.Sprintf("article-%d", i), Status: "published", PublishedAt: now, Category: "go", Tags: []string{"testing"}})
		}
		h := articles.NewHandler(query.NewEditorial(catalogue{all}, func() time.Time { return now }), markdown.NewRenderer())
		for _, page := range []int{1, 2} {
			w := httptest.NewRecorder()
			ctx := gohttp.NewContext(w, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/articles?category=go&tag=testing&page=%d", page), nil))
			ctx.SetRenderer(gohttp.NewReloadingHTMLRenderer(filepath.Join("..", "..", "..", "..", "tmpl")))
			assert.NilError(t, h.List(ctx, struct{}{}))
			if page == 1 {
				assert.Assert(t, strings.Contains(w.Body.String(), `/articles?category=go&amp;page=2&amp;tag=testing`))
			} else {
				assert.Assert(t, strings.Contains(w.Body.String(), `/articles?category=go&amp;tag=testing`))
				assert.Assert(t, !strings.Contains(w.Body.String(), `rel="next"`))
			}
		}
	})
}

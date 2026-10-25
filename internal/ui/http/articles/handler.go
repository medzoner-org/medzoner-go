package articles

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Medzoner/medzoner-go/internal/application/query"
	"github.com/Medzoner/medzoner-go/internal/domains/editorial"
	port "github.com/Medzoner/medzoner-go/internal/ports/editorial"
	gohttp "github.com/medzoner-org/gomedz/pkg/http"
	"github.com/medzoner-org/gomedz/pkg/observability"
)

const baseURL = "https://www.medzoner.com"

type View struct {
	Locale          string
	PageTitle       string
	PageDescription string
	TorHost         string
	CanonicalURL    string
	SocialImage     string
	Article         editorial.Article
	Body            template.HTML
	StructuredData  template.JS
	Headings        []port.Heading
}

type Handler struct {
	queries  *query.Editorial
	renderer port.Renderer
}

func NewHandler(queries *query.Editorial, renderer port.Renderer) Handler {
	return Handler{queries: queries, renderer: renderer}
}

func (Handler) Prefix() string { return "/" }

func (h Handler) Register(r gohttp.Router[any]) {
	r.Get("/articles", h.List, gohttp.Options{})
	r.Get("/articles/:slug", h.Show, gohttp.Options{})
	r.Get("/feed.xml", h.Feed, gohttp.Options{})
	r.Get("/sitemap.xml", h.Sitemap, gohttp.Options{})
}

func (h Handler) Show(c *gohttp.Context, _ struct{}) error {
	ctx, span := observability.StartSpan(c.Context(), "Articles.Show")
	defer span.End()
	a, err := h.queries.Get(ctx, c.Param("slug"))
	if errors.Is(err, editorial.ErrNotFound) {
		return c.HTML(http.StatusNotFound, "404", nil)
	}
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("query article: %w", err)
	}
	body, err := h.renderer.Render(ctx, a.Markdown)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("render markdown article: %w", err)
	}
	canonical := fmt.Sprintf("%s/articles/%s", baseURL, a.Slug)
	socialImage := fmt.Sprintf("%s%s", baseURL, a.Image)
	schema := map[string]any{
		"@context": "https://schema.org", "@type": "TechArticle",
		"headline": a.Title, "description": a.Description,
		"author":           map[string]string{"@type": "Person", "name": a.Author.Name, "url": a.Author.URL},
		"mainEntityOfPage": canonical, "image": socialImage, "inLanguage": a.Language,
		"datePublished": a.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if !a.UpdatedAt.IsZero() {
		schema["dateModified"] = a.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	data, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("encode article schema: %w", err)
	}
	view := View{
		Locale: a.Language, PageTitle: fmt.Sprintf("%s | Medzoner", a.Title), PageDescription: a.Description,
		CanonicalURL: canonical, SocialImage: socialImage, Article: a,
		// Only the safe Markdown renderer output crosses this trust boundary.
		Body:     template.HTML(body.HTML), //nolint:gosec // Goldmark has raw HTML and unsafe links disabled.
		Headings: body.Headings,
		// json.Marshal escapes HTML characters, including closing script sequences.
		StructuredData: template.JS(data), //nolint:gosec // JSON generated from typed editorial metadata, not request input.
	}
	if err := c.HTML(http.StatusOK, "article", view); err != nil {
		span.RecordError(err)
		return fmt.Errorf("render article: %w", err)
	}
	return nil
}

type ListView struct {
	Locale       string
	TorHost      string
	CanonicalURL string
	Category     string
	Tag          string
	Result       query.ArticlePage
	PreviousURL  string
	NextURL      string
}

func listURL(page int, category, tag string) string {
	values := url.Values{}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	if category != "" {
		values.Set("category", category)
	}
	if tag != "" {
		values.Set("tag", tag)
	}
	if len(values) == 0 {
		return "/articles"
	}
	return fmt.Sprintf("/articles?%s", values.Encode())
}

func (h Handler) List(c *gohttp.Context, _ struct{}) error {
	ctx, span := observability.StartSpan(c.Context(), "Articles.List")
	defer span.End()
	values := c.Request().URL.Query()
	page := 1
	if values.Has("page") {
		var err error
		page, err = strconv.Atoi(values.Get("page"))
		if err != nil || page < 1 {
			http.Error(c.Writer(), "Invalid page", http.StatusBadRequest)
			return nil
		}
	}
	filter := query.ListArticles{Page: page, PageSize: 10, Category: values.Get("category"), Tag: values.Get("tag")}
	result, err := h.queries.List(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("list articles: %w", err)
	}
	view := ListView{Locale: "fr", Result: result, Category: filter.Category, Tag: filter.Tag, CanonicalURL: fmt.Sprintf("%s%s", baseURL, listURL(page, filter.Category, filter.Tag))}
	if page > 1 {
		view.PreviousURL = listURL(page-1, filter.Category, filter.Tag)
	}
	if page <= result.Total/result.PageSize && page*result.PageSize < result.Total {
		view.NextURL = listURL(page+1, filter.Category, filter.Tag)
	}
	if err := c.HTML(http.StatusOK, "article_list", view); err != nil {
		span.RecordError(err)
		return fmt.Errorf("render article list: %w", err)
	}
	return nil
}

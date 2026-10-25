package articles

import (
	"encoding/xml"
	"fmt"
	"time"

	gohttp "github.com/medzoner-org/gomedz/pkg/http"
	"github.com/medzoner-org/gomedz/pkg/observability"
)

type sitemapURL struct {
	Location string `xml:"loc"`
	Modified string `xml:"lastmod,omitempty"`
}

type sitemap struct {
	XMLName   xml.Name     `xml:"urlset"`
	Namespace string       `xml:"xmlns,attr"`
	URLs      []sitemapURL `xml:"url"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	Published   string `xml:"pubDate"`
	Category    string `xml:"category"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

func writeXML(c *gohttp.Context, contentType string, value any) error {
	data, err := xml.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode editorial XML: %w", err)
	}
	c.SetHeader("Content-Type", contentType)
	if _, err := fmt.Fprint(c.Writer(), xml.Header, string(data)); err != nil {
		return fmt.Errorf("write editorial XML: %w", err)
	}
	return nil
}

func (h Handler) Sitemap(c *gohttp.Context, _ struct{}) error {
	ctx, span := observability.StartSpan(c.Context(), "Articles.Sitemap")
	defer span.End()
	articles, err := h.queries.Published(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("sitemap articles: %w", err)
	}
	doc := sitemap{Namespace: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: []sitemapURL{{Location: baseURL + "/"}, {Location: baseURL + "/articles"}}}
	for _, a := range articles {
		modified := a.PublishedAt
		if !a.UpdatedAt.IsZero() {
			modified = a.UpdatedAt
		}
		doc.URLs = append(doc.URLs, sitemapURL{Location: fmt.Sprintf("%s/articles/%s", baseURL, a.Slug), Modified: modified.Format("2006-01-02")})
	}
	return writeXML(c, "application/xml; charset=utf-8", doc)
}

func (h Handler) Feed(c *gohttp.Context, _ struct{}) error {
	ctx, span := observability.StartSpan(c.Context(), "Articles.Feed")
	defer span.End()
	articles, err := h.queries.Published(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("feed articles: %w", err)
	}
	doc := rss{Version: "2.0", Channel: rssChannel{Title: "Medzoner", Link: baseURL, Description: "Go, DevOps et IoT : articles techniques et projets open source."}}
	for _, a := range articles[:min(20, len(articles))] {
		link := fmt.Sprintf("%s/articles/%s", baseURL, a.Slug)
		doc.Channel.Items = append(doc.Channel.Items, rssItem{Title: a.Title, Link: link, GUID: link, Description: a.Description, Published: a.PublishedAt.Format(time.RFC1123Z), Category: a.Category})
	}
	return writeXML(c, "application/rss+xml; charset=utf-8", doc)
}

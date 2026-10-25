// Package editorial loads versioned Markdown. It never reads a request-controlled path.
package editorial

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"regexp"
	"strings"

	"github.com/Medzoner/medzoner-go/internal/domains/editorial"
	port "github.com/Medzoner/medzoner-go/internal/ports/editorial"
	"github.com/medzoner-org/gomedz/pkg/observability"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Repository struct{ articles []editorial.Article }

// NewRepository validates all content once at startup, including drafts.
func NewRepository(ctx context.Context, source fs.FS) (_ *Repository, err error) {
	ctx, span := observability.StartSpan(ctx, "MarkdownRepository.New")
	defer span.End()
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()
	files, err := fs.Glob(source, "*.md")
	if err != nil {
		return nil, fmt.Errorf("find editorial files: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("empty editorial catalogue")
	}
	repo := &Repository{}
	seen := make(map[string]bool)
	for _, name := range files {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("load catalogue: %w", err)
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return nil, fmt.Errorf("read article %s: %w", name, err)
		}
		a, err := parse(data)
		if err != nil {
			return nil, fmt.Errorf("parse article %s: %w", name, err)
		}
		if seen[a.Slug] {
			return nil, fmt.Errorf("duplicate article slug %s", a.Slug)
		}
		seen[a.Slug] = true
		repo.articles = append(repo.articles, a)
	}
	return repo, nil
}

func (r *Repository) All(ctx context.Context) ([]editorial.Article, error) {
	ctx, span := observability.StartSpan(ctx, "MarkdownRepository.All")
	defer span.End()
	if err := ctx.Err(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("read catalogue: %w", err)
	}
	result := make([]editorial.Article, len(r.articles))
	for i, a := range r.articles {
		a.Tags = append([]string(nil), a.Tags...)
		result[i] = a
	}
	return result, nil
}

func parse(data []byte) (editorial.Article, error) {
	parts := bytes.SplitN(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")), []byte("\n---\n"), 2)
	if len(parts) != 2 || !bytes.HasPrefix(parts[0], []byte("---\n")) {
		return editorial.Article{}, fmt.Errorf("expected YAML front matter")
	}
	var a editorial.Article
	decoder := yaml.NewDecoder(bytes.NewReader(parts[0][4:]))
	decoder.KnownFields(true)
	if err := decoder.Decode(&a); err != nil {
		return a, fmt.Errorf("decode metadata: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return a, fmt.Errorf("expected one metadata document")
	}
	a.Markdown = strings.TrimSpace(string(parts[1]))
	if err := validate(a); err != nil {
		return a, err
	}
	return a, nil
}

func validate(a editorial.Article) error {
	if !slugPattern.MatchString(a.Slug) || !slugPattern.MatchString(a.Category) {
		return fmt.Errorf("invalid slug or category")
	}
	if strings.TrimSpace(a.Title) == "" || strings.TrimSpace(a.Description) == "" || strings.TrimSpace(a.Author.Name) == "" || a.Markdown == "" {
		return fmt.Errorf("title, description, author and body are required")
	}
	if a.Language != "en" && a.Language != "fr" {
		return fmt.Errorf("language must be en or fr")
	}
	if a.Status != "draft" && a.Status != "published" {
		return fmt.Errorf("status must be draft or published")
	}
	if a.Status == "published" && a.PublishedAt.IsZero() {
		return fmt.Errorf("published article needs published_at")
	}
	if !a.UpdatedAt.IsZero() && (a.PublishedAt.IsZero() || a.UpdatedAt.Before(a.PublishedAt)) {
		return fmt.Errorf("updated_at must not precede published_at")
	}
	if !strings.HasPrefix(a.Image, "/public/images/") || strings.ContainsAny(a.Image, "?#\\") || strings.Contains(a.Image, "..") {
		return fmt.Errorf("image must be a local public image")
	}
	u, err := url.Parse(a.Author.URL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("author URL must be absolute HTTPS")
	}
	for _, tag := range a.Tags {
		if !slugPattern.MatchString(tag) {
			return fmt.Errorf("invalid tag %q", tag)
		}
	}
	return nil
}

type Renderer struct{ markdown goldmark.Markdown }

func NewRenderer() *Renderer {
	// No WithUnsafe: raw HTML and dangerous link protocols remain disabled.
	return &Renderer{markdown: goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID(), parser.WithAttribute()),
		goldmark.WithRendererOptions(renderer.WithNodeRenderers(util.Prioritized(editorialNodes{}, 500))),
	)}
}

func (r *Renderer) Render(ctx context.Context, markdown string) (port.Document, error) {
	ctx, span := observability.StartSpan(ctx, "MarkdownRenderer.Render")
	defer span.End()
	if err := ctx.Err(); err != nil {
		return port.Document{}, fmt.Errorf("render markdown: %w", err)
	}
	source := []byte(markdown)
	document := r.markdown.Parser().Parse(text.NewReader(source))
	result := port.Document{}
	if err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Authored attributes are never trusted, except safe heading identifiers.
		id, hasID := node.AttributeString("id")
		node.RemoveAttributes()
		if heading, ok := node.(*ast.Heading); ok && hasID {
			if value, ok := id.([]byte); ok && anchorPattern.Match(value) {
				node.SetAttributeString("id", value)
				if heading.Level == 2 {
					result.Headings = append(result.Headings, port.Heading{ID: string(value), Title: inlineText(node, source)})
				}
			}
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return port.Document{}, fmt.Errorf("validate Markdown AST: %w", err)
	}
	prepareCallouts(document, source)
	var out bytes.Buffer
	if err := r.markdown.Renderer().Render(&out, source, document); err != nil {
		return port.Document{}, fmt.Errorf("render markdown: %w", err)
	}
	result.HTML = out.String()
	return result, nil
}

var anchorPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

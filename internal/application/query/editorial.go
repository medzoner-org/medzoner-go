package query

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/Medzoner/medzoner-go/internal/domains/editorial"
	port "github.com/Medzoner/medzoner-go/internal/ports/editorial"
	"github.com/medzoner-org/gomedz/pkg/observability"
)

type Editorial struct {
	repository port.Repository
	now        func() time.Time
}

func NewEditorial(repository port.Repository, now func() time.Time) *Editorial {
	return &Editorial{repository: repository, now: now}
}

type ListArticles struct {
	Category string
	Tag      string
	Page     int
	PageSize int
}

type ArticlePage struct {
	Articles []editorial.Article
	Total    int
	Page     int
	PageSize int
}

func (q *Editorial) Published(ctx context.Context) ([]editorial.Article, error) {
	ctx, span := observability.StartSpan(ctx, "Editorial.Published")
	defer span.End()
	all, err := q.repository.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load editorial catalogue: %w", err)
	}
	articles := make([]editorial.Article, 0, len(all))
	now := q.now()
	for _, a := range all {
		if a.IsPublic(now) {
			articles = append(articles, a)
		}
	}
	sort.Slice(articles, func(i, j int) bool {
		if articles[i].PublishedAt.Equal(articles[j].PublishedAt) {
			return articles[i].Slug < articles[j].Slug
		}
		return articles[i].PublishedAt.After(articles[j].PublishedAt)
	})
	return articles, nil
}

func (q *Editorial) Get(ctx context.Context, slug string) (editorial.Article, error) {
	ctx, span := observability.StartSpan(ctx, "Editorial.Get")
	defer span.End()
	articles, err := q.Published(ctx)
	if err != nil {
		return editorial.Article{}, err
	}
	for _, a := range articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return editorial.Article{}, fmt.Errorf("get article: %w", editorial.ErrNotFound)
}

func (q *Editorial) List(ctx context.Context, filter ListArticles) (ArticlePage, error) {
	ctx, span := observability.StartSpan(ctx, "Editorial.List")
	defer span.End()
	if filter.Page < 1 || filter.PageSize < 1 || filter.PageSize > 50 {
		return ArticlePage{}, fmt.Errorf("pagination: %w", editorial.ErrInvalidQuery)
	}
	articles, err := q.Published(ctx)
	if err != nil {
		return ArticlePage{}, err
	}
	filtered := make([]editorial.Article, 0, len(articles))
	for _, a := range articles {
		if (filter.Category == "" || a.Category == filter.Category) && (filter.Tag == "" || slices.Contains(a.Tags, filter.Tag)) {
			filtered = append(filtered, a)
		}
	}
	result := ArticlePage{Total: len(filtered), Page: filter.Page, PageSize: filter.PageSize}
	// Compare before multiplying to avoid overflow on arbitrary page parameters.
	if len(filtered) == 0 || filter.Page > (len(filtered)-1)/filter.PageSize+1 {
		return result, nil
	}
	start := (filter.Page - 1) * filter.PageSize
	result.Articles = filtered[start:min(start+filter.PageSize, len(filtered))]
	return result, nil
}

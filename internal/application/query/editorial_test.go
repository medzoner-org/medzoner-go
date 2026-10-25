package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Medzoner/medzoner-go/internal/application/query"
	"github.com/Medzoner/medzoner-go/internal/domains/editorial"
	"gotest.tools/assert"
)

type repository struct {
	articles []editorial.Article
	err      error
}

func (r repository) All(context.Context) ([]editorial.Article, error) { return r.articles, r.err }

func TestEditorial(t *testing.T) {
	now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	q := query.NewEditorial(repository{articles: []editorial.Article{
		{Slug: "older", Status: "published", PublishedAt: now.Add(-time.Hour), Category: "go", Tags: []string{"testing"}},
		{Slug: "draft", Status: "draft", PublishedAt: now},
		{Slug: "scheduled", Status: "published", PublishedAt: now.Add(time.Second)},
		{Slug: "newer", Status: "published", PublishedAt: now, Category: "iot"},
	}}, func() time.Time { return now })
	t.Run("Unit: test only public articles sorted newest first", func(t *testing.T) {
		got, err := q.Published(context.Background())
		assert.NilError(t, err)
		assert.Equal(t, len(got), 2)
		assert.Equal(t, got[0].Slug, "newer")
		assert.Equal(t, got[1].Slug, "older")
	})
	t.Run("Unit: test drafts scheduled and unknown slugs remain not found", func(t *testing.T) {
		for _, slug := range []string{"draft", "scheduled", "unknown", "../older"} {
			_, err := q.Get(context.Background(), slug)
			assert.Assert(t, errors.Is(err, editorial.ErrNotFound))
		}
	})
	t.Run("Unit: test pagination and combined filters", func(t *testing.T) {
		got, err := q.List(context.Background(), query.ListArticles{Page: 2, PageSize: 1})
		assert.NilError(t, err)
		assert.Equal(t, got.Total, 2)
		assert.Equal(t, got.Articles[0].Slug, "older")
		got, err = q.List(context.Background(), query.ListArticles{Page: 1, PageSize: 10, Category: "go", Tag: "testing"})
		assert.NilError(t, err)
		assert.Equal(t, got.Total, 1)
		got, err = q.List(context.Background(), query.ListArticles{Page: int(^uint(0) >> 1), PageSize: 10})
		assert.NilError(t, err)
		assert.Equal(t, len(got.Articles), 0)
	})
	t.Run("Unit: test rejects invalid pagination", func(t *testing.T) {
		for _, filter := range []query.ListArticles{{Page: 0, PageSize: 10}, {Page: 1, PageSize: 51}, {Page: 1, PageSize: -1}} {
			_, err := q.List(context.Background(), filter)
			assert.Assert(t, errors.Is(err, editorial.ErrInvalidQuery))
		}
	})
	t.Run("Unit: test repository error stays wrapped", func(t *testing.T) {
		cause := errors.New("storage unavailable")
		failed := query.NewEditorial(repository{err: cause}, func() time.Time { return now })
		_, err := failed.Published(context.Background())
		assert.Assert(t, errors.Is(err, cause))
	})
}

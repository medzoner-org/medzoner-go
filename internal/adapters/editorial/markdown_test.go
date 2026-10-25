package editorial_test

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	markdown "github.com/Medzoner/medzoner-go/internal/adapters/editorial"
	"github.com/Medzoner/medzoner-go/internal/resources"
	"gotest.tools/assert"
)

const article = `---
slug: test-article
title: Test article
description: A description
author:
  name: Author
  url: https://www.medzoner.com/
category: go
tags: [go]
language: en
image: /public/images/whyor-social.png
status: published
published_at: 2026-10-07T20:55:54+02:00
---
## Heading {#stable-id}

A paragraph with **Markdown**.
`

func source(content string) fstest.MapFS {
	return fstest.MapFS{"article.md": &fstest.MapFile{Data: []byte(content)}}
}

func TestRepository(t *testing.T) {
	t.Run("Unit: test parses metadata and returns defensive copies", func(t *testing.T) {
		repo, err := markdown.NewRepository(context.Background(), source(article))
		assert.NilError(t, err)
		got, err := repo.All(context.Background())
		assert.NilError(t, err)
		assert.Equal(t, got[0].Slug, "test-article")
		assert.Assert(t, strings.Contains(got[0].Markdown, "## Heading"))
		got[0].Tags[0] = "changed"
		again, err := repo.All(context.Background())
		assert.NilError(t, err)
		assert.Equal(t, again[0].Tags[0], "go")
	})
	t.Run("Unit: test invalid catalogue fails at startup", func(t *testing.T) {
		for _, invalid := range []string{
			"No metadata", strings.Replace(article, "slug: test-article", "slug: ../escape", 1),
			strings.Replace(article, "status: published", "status: unknown", 1),
			strings.Replace(article, "title: Test article", "title:", 1),
			strings.Replace(article, "language: en", "language: unknown", 1),
			strings.Replace(article, "category: go", "category: ../invalid", 1),
			strings.Replace(article, "image: /public/images/whyor-social.png", "image: javascript:alert(1)", 1),
			strings.Replace(article, "published_at: 2026-10-07T20:55:54+02:00", "published_at: invalid", 1),
			strings.Replace(article, "published_at: 2026-10-07T20:55:54+02:00", "", 1),
			strings.Replace(article, "title: Test article", "unknown_field: test", 1),
			strings.Replace(article, "url: https://www.medzoner.com/", "url: javascript:alert(1)", 1),
			strings.Replace(article, "tags: [go]", "tags: [../invalid]", 1),
			strings.Replace(article, "status: published", "status: published\nupdated_at: 2020-01-01T00:00:00Z", 1),
		} {
			_, err := markdown.NewRepository(context.Background(), source(invalid))
			assert.Assert(t, err != nil, "invalid input unexpectedly accepted: %s", invalid)
		}
		duplicate := source(article)
		duplicate["duplicate.md"] = &fstest.MapFile{Data: []byte(article)}
		_, err := markdown.NewRepository(context.Background(), duplicate)
		assert.ErrorContains(t, err, "duplicate")
		_, err = markdown.NewRepository(context.Background(), fstest.MapFS{})
		assert.ErrorContains(t, err, "empty")
	})
	t.Run("Unit: test embedded articles validate independently of working directory", func(t *testing.T) {
		files, err := resources.EditorialFiles()
		assert.NilError(t, err)
		repo, err := markdown.NewRepository(context.Background(), files)
		assert.NilError(t, err)
		got, err := repo.All(context.Background())
		assert.NilError(t, err)
		assert.Assert(t, len(got) >= 2)
	})
	t.Run("Unit: test cancelled context is propagated", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := markdown.NewRepository(ctx, source(article))
		assert.ErrorContains(t, err, "context canceled")
	})
}

func TestSafeRenderer(t *testing.T) {
	t.Run("Unit: test Markdown supports code and stable anchors but not executable HTML", func(t *testing.T) {
		got, err := markdown.NewRenderer().Render(context.Background(), "## Heading {#stable-id}\n\n<script>alert(1)</script>\n\n[bad](javascript:alert(1))\n\n![bad](data:text/html,evil)\n\n```go\nvar x = 1\n```")
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(got.HTML, `id="stable-id"`))
		assert.Assert(t, strings.Contains(got.HTML, "var x = 1"))
		for _, unsafe := range []string{"<script>", "javascript:", "data:text/html"} {
			assert.Assert(t, !strings.Contains(got.HTML, unsafe), "unsafe output: %s", got.HTML)
		}
	})
	t.Run("Unit: test heading attributes are limited to safe anchors", func(t *testing.T) {
		got, err := markdown.NewRenderer().Render(context.Background(), "## Heading {#stable-id style=\"position:fixed;inset:0\" contenteditable=\"true\" autofocus=\"true\" data-secret=\"yes\" onclick=\"alert(1)\"}\n")
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(got.HTML, `id="stable-id"`))
		for _, attribute := range []string{"style=", "contenteditable", "autofocus", "data-secret", "onclick"} {
			assert.Assert(t, !strings.Contains(got.HTML, attribute), "unexpected attribute: %s", got.HTML)
		}
		assert.Equal(t, len(got.Headings), 1)
		assert.Equal(t, got.Headings[0].ID, "stable-id")
	})
	t.Run("Unit: test safe notes captions and call to action presentation", func(t *testing.T) {
		got, err := markdown.NewRenderer().Render(context.Background(), "> **Note:** a boundary.\n\n![Alt](/public/images/example.png \"A caption\")\n\n> [!CTA]\n>\n> [Explore](https://example.com/)\n")
		assert.NilError(t, err)
		for _, want := range []string{`<aside class="article-note">`, `<figure class="article-figure">`, "<figcaption>A caption</figcaption>", `<div class="article-cta">`, `class="btn btn-primary"`} {
			assert.Assert(t, strings.Contains(got.HTML, want), "missing preserved presentation: %s; got %s", want, got.HTML)
		}
		assert.Assert(t, !strings.Contains(got.HTML, "[!CTA]"))
	})
}

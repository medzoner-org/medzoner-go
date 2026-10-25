// Package editorial defines the publication model independently of storage and HTTP.
package editorial

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("article not found")
	ErrInvalidQuery = errors.New("invalid editorial query")
)

type Author struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Article struct {
	Slug        string    `yaml:"slug"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Deck        string    `yaml:"deck"`
	Author      Author    `yaml:"author"`
	Category    string    `yaml:"category"`
	Tags        []string  `yaml:"tags"`
	Language    string    `yaml:"language"`
	Image       string    `yaml:"image"`
	ImageAlt    string    `yaml:"image_alt"`
	Status      string    `yaml:"status"`
	PublishedAt time.Time `yaml:"published_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
	Markdown    string    `yaml:"-"`
}

func (a Article) IsPublic(now time.Time) bool {
	return a.Status == "published" && !a.PublishedAt.IsZero() && !a.PublishedAt.After(now)
}

package editorial

import "context"

type Heading struct {
	ID    string
	Title string
}

type Document struct {
	HTML     string
	Headings []Heading
}

type Renderer interface {
	Render(context.Context, string) (Document, error)
}

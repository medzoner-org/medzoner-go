package editorial

import (
	"html"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// prepareCallouts adds only application-owned presentation classes after scrubbing
// all authored attributes. No arbitrary class/style/HTML directive is supported.
func prepareCallouts(document ast.Node, source []byte) {
	for node := document.FirstChild(); node != nil; node = node.NextSibling() {
		if node.Kind() != ast.KindBlockquote {
			continue
		}
		first := node.FirstChild()
		if first == nil || inlineText(first, source) != "[!CTA]" {
			continue
		}
		node.SetAttributeString("editorial-cta", true)
		node.RemoveChild(node, first)
		for paragraph := node.FirstChild(); paragraph != nil; paragraph = paragraph.NextSibling() {
			link, ok := paragraph.FirstChild().(*ast.Link)
			if ok && link.NextSibling() == nil {
				link.SetAttributeString("class", []byte("btn btn-primary"))
			}
		}
	}
}

func inlineText(node ast.Node, source []byte) string {
	var out strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch value := child.(type) {
		case *ast.Text:
			out.Write(value.Value(source))
		case *ast.String:
			out.Write(value.Value)
		default:
			out.WriteString(inlineText(child, source))
		}
	}
	return html.UnescapeString(out.String())
}

type editorialNodes struct{}

func (editorialNodes) RegisterFuncs(register renderer.NodeRendererFuncRegisterer) {
	register.Register(ast.KindBlockquote, renderCallout)
	register.Register(ast.KindParagraph, renderParagraph)
}

func renderCallout(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	_, cta := node.AttributeString("editorial-cta")
	if entering {
		if cta {
			_, _ = w.WriteString("<div class=\"article-cta\">\n")
		} else {
			_, _ = w.WriteString("<aside class=\"article-note\">\n")
		}
	} else if cta {
		_, _ = w.WriteString("</div>\n")
	} else {
		_, _ = w.WriteString("</aside>\n")
	}
	return ast.WalkContinue, nil
}

func renderParagraph(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	image, figure := node.FirstChild().(*ast.Image)
	figure = figure && image.NextSibling() == nil
	if !figure {
		if entering {
			_, _ = w.WriteString("<p>")
		} else {
			_, _ = w.WriteString("</p>\n")
		}
		return ast.WalkContinue, nil
	}
	if entering {
		_, _ = w.WriteString("<figure class=\"article-figure\">")
	} else {
		if len(image.Title) > 0 {
			_, _ = w.WriteString("<figcaption>")
			_, _ = w.WriteString(html.EscapeString(string(image.Title)))
			_, _ = w.WriteString("</figcaption>")
		}
		_, _ = w.WriteString("</figure>\n")
	}
	return ast.WalkContinue, nil
}

// Package koluszki implements Renderer to render
// github.com/maragudk/gomponents Go code based on html.Node. It also provides
// CLI and HTTP server for your convenience.
package koluszki

import (
	"io"

	"golang.org/x/net/html"
)

// Render renders given html.Node into provided writer. Use [NewRenderer] if
// you need more control over output.
func Render(w io.Writer, n *html.Node) error {
	r := NewRenderer(
		WithGomponentsAlias("g"),
		WithHTMLPackageElements("g", ""),
		WithHTMLPackageAttributes("g", ""),
		WithRenderSVG,
	)

	return r.Render(w, n)
}

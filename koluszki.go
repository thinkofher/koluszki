package koluszki

import (
	"io"

	"golang.org/x/net/html"
)

func Render(w io.Writer, n *html.Node) error {
	r := NewRenderer(
		WithGomponentsAlias("g"),
		WithHTMLPackageElements("g", "h"),
		WithHTMLPackageAttributes("g", "h"),
		WithRenderSVG,
	)
	return r.Render(w, n)
}

package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/net/html"

	"github.com/thinkofher/koluszki"
)

func findFirstChild(n *html.Node) *html.Node {
	if n.Type == html.DocumentNode {
		return findFirstChild(n.FirstChild)
	}

	if n.Type == html.CommentNode {
		if n.NextSibling != nil {
			return findFirstChild(n.NextSibling)
		}
	}

	return n
}

func run() error {
	var (
		gomponentsAlias string
		htmlAlias       string
		htmlEnabled     bool
		svgEnabled      bool
	)

	flag.StringVar(&gomponentsAlias, "gomponentsAlias", "g", "alias for gomponents package")
	flag.StringVar(&htmlAlias, "htmlAlias", "", "alias for gomponents/html package")
	flag.BoolVar(&htmlEnabled, "htmlEnabled", true, "render elements from gomponents/html package")
	flag.BoolVar(&svgEnabled, "svg", false, "if true, koluszki will fully render svg elements")
	flag.Parse()

	n, err := html.Parse(os.Stdin)
	if err != nil {
		return fmt.Errorf("parse stream: %w", err)
	}

	opts := []koluszki.Option{}

	if gomponentsAlias != "" {
		opts = append(opts, koluszki.WithGomponentsAlias(gomponentsAlias))
	}
	if htmlEnabled {
		opts = append(opts,
			koluszki.WithHTMLPackageAttributes(gomponentsAlias, htmlAlias),
			koluszki.WithHTMLPackageElements(gomponentsAlias, htmlAlias),
		)
	}
	if svgEnabled {
		opts = append(opts, koluszki.WithRenderSVG)
	}

	renderer := koluszki.NewRenderer(opts...)
	if err := renderer.Render(os.Stdout, findFirstChild(n)); err != nil {
		return fmt.Errorf("koluszki render: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

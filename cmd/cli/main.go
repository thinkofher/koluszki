package main

import (
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
	n, err := html.Parse(os.Stdin)
	if err != nil {
		return fmt.Errorf("parse stream: %w", err)
	}

	if err := koluszki.Render(os.Stdout, findFirstChild(n)); err != nil {
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

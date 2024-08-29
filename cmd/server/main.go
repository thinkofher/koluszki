package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"

	"github.com/thinkofher/koluszki"
)

//go:embed index.html
var index string

var (
	tmplIndex *template.Template
	tmplErr   error
	tmplOnce  sync.Once
)

type tmplProperties struct {
	Code string
}

const tmplDefaultCode = `g.Text("There will be golang here...")`

func indexTemplate() (*template.Template, error) {
	tmplOnce.Do(func() {
		tmplIndex, tmplErr = template.New("tmpl").Parse(index)
	})

	return tmplIndex, tmplErr
}

func indexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := indexTemplate()
		if err != nil {
			http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
			return
		}

		props := tmplProperties{
			Code: tmplDefaultCode,
		}

		if err := tmpl.ExecuteTemplate(w, "index", &props); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func codeHandler() http.Handler {
	type rendererPropersties struct {
		renderSVG       bool
		useHTMLPackage  bool
		htmlAlias       string
		gomponentsAlias string
	}

	renderer := func(props rendererPropersties) *koluszki.Renderer {
		opts := []koluszki.Option{}

		if props.gomponentsAlias != "" {
			opts = append(opts, koluszki.WithGomponentsAlias(props.gomponentsAlias))
		}

		if props.useHTMLPackage {
			opts = append(opts,
				koluszki.WithHTMLPackageAttributes(props.gomponentsAlias, props.htmlAlias),
				koluszki.WithHTMLPackageElements(props.gomponentsAlias, props.htmlAlias),
			)
		}

		if props.renderSVG {
			opts = append(opts, koluszki.WithRenderSVG)
		}

		return koluszki.NewRenderer(opts...)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := indexTemplate()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
			return
		}

		props := tmplProperties{
			Code: tmplDefaultCode,
		}

		code := r.Form.Get("code")
		if code == "" {
			if err := tmpl.ExecuteTemplate(w, "code", &props); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		in := strings.NewReader(code)
		nodes, err := html.Parse(in)
		if err != nil {
			http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
			return
		}

		rr := renderer(rendererPropersties{
			renderSVG:       r.Form.Get("svg") != "",
			useHTMLPackage:  r.Form.Get("html-enabled") != "",
			htmlAlias:       r.Form.Get("html-pkg"),
			gomponentsAlias: r.Form.Get("gomponents"),
		})

		var buff strings.Builder
		if err := rr.Render(&buff, nodes.FirstChild); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// I am sorry, but I'm just lazy. For some reason there is newline rendered
		// on the first line. Let's just trim it and call it a day.
		props.Code = strings.TrimLeft(buff.String(), "\n")

		if err := tmpl.ExecuteTemplate(w, "code", &props); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func run() error {
	var (
		host string
		port int
	)

	flag.StringVar(&host, "host", "127.0.0.1", "listen host address")
	flag.IntVar(&port, "port", 21037, "listen port")
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("GET /{$}", indexHandler())
	mux.Handle("POST /code", codeHandler())

	addr := net.JoinHostPort(host, strconv.Itoa(port))

	fmt.Printf("Listening at %s...", addr)
	return http.ListenAndServe(addr, mux)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

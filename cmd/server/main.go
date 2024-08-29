package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
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

		var buff strings.Builder
		if err := koluszki.Render(&buff, nodes.FirstChild); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		props.Code = buff.String()
		if err := tmpl.ExecuteTemplate(w, "code", &props); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func run() error {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", indexHandler())
	mux.Handle("POST /code", codeHandler())

	fmt.Println("Listening at 127.0.0.1:21037...")
	return http.ListenAndServe("127.0.0.1:21037", mux)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

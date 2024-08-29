# Koluszki

Koluszki is a Go package that transforms HTML code into Go code that renders
provided HTML with amazing [gomponents](https://github.com/maragudk/gomponents)
library.

## Usage

Koluszki provides [library](https://pkg.go.dev/github.com/thinkofher/koluszki), CLI and HTTP server to translate HTML into Go code.

### CLI

You can use `cmd/cli` program to translate HTML code from standard input to a
go code that is going to be streamed into standard output.

```sh
$ echo '<h1>Hello World!</h1>' | go run github.com/thinkofher/koluszki/cmd/cli@latest
HTML(
 Head(
 ),
 Body(
  H1(
   g.Text(`Hello World!`),
  ),
 ),
)
```

The `cmd/cli` package integrates well with other Unix tools.

```sh
luxemburg; curl -s 'https://github.com' | go run github.com/thinkofher/koluszki/cmd/cli@latest | head -n 20
HTML(
 Lang("en"),
 Data("color-mode", "light"),
 Data("light-theme", "light"),
 Data("dark-theme", "dark"),
 Data("a11y-animated-images", "system"),
 Data("a11y-link-underlines", "true"),
 Head(
  Meta(
   Charset("utf-8"),
  ),
  Link(
   Rel("dns-prefetch"),
   Href("https://github.githubassets.com"),
  ),
  Link(
   Rel("dns-prefetch"),
   Href("https://avatars.githubusercontent.com"),
  ),
```

### HTTP Server

Just execute below command in the terminal and visit the prompted address in
the browser. You'll see a web application that executes Koluszki's interface
through a HTTP integration.

```sh
$ go run github.com/thinkofher/koluszki/cmd/server@latest --host localhost --port 8080
Listening at localhost:8080...
```

## Extras

You can read about beautiful polish city Koluszki [here](https://en.wikipedia.org/wiki/Koluszki).

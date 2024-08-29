package koluszki

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Renderer renders github.com/maragudk/gomponents go code.
//
// Use [NewRenderer] to initialize [Renderer]. Using default value of Renderer
// can end up with panic.
type Renderer struct {
	renderSVG         bool
	attrWrite         func(w io.Writer, level int, attr html.Attribute) error
	nodeWrite         func(w io.Writer, level int, tag string) error
	gomponentsPkgName string
}

// NewRenderer returns new Renderer. It is the only way to create new intance
// of [Renderer].
func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{
		renderSVG:         false,
		gomponentsPkgName: "gomponents",
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.nodeWrite == nil {
		r.nodeWrite = rawElementsWriter(r.gomponentsPkgName)
	}

	if r.attrWrite == nil {
		r.attrWrite = rawAttributesWriter(r.gomponentsPkgName)
	}

	return r
}

// Render renders given html.Node into provided writer.
func (r *Renderer) Render(w io.Writer, n *html.Node) error {
	return r.render(w, n, 0)
}

func (r *Renderer) render(w io.Writer, n *html.Node, level int) error {
	pref := "\n"
	prefAttr := "\n"

	if level != 0 {
		for range level {
			pref += " "
		}

		for range level + 1 {
			prefAttr += " "
		}
	}

	skip := true
	isSVG := false

	switch {
	case n.Type == html.TextNode:
		if n.Data != "" && strings.ContainsAny(n.Data, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") {
			fmt.Fprintf(w, "%s%s.Text(`%s`),", pref, r.gomponentsPkgName, strings.Trim(n.Data, "\n"))
		}
	case !r.renderSVG && n.Data == "svg":
		var buff strings.Builder
		html.Render(&buff, n)
		fmt.Fprintf(w, "%s%s.Raw(`%s`),", pref, r.gomponentsPkgName, buff.String())
		skip = false
		isSVG = true
	case n.Type == html.ElementNode:
		r.nodeWrite(w, level, n.Data)
		skip = false
	}

	if !isSVG {
		for _, a := range n.Attr {
			r.attrWrite(w, level, a)
		}

		if next := n.FirstChild; next != nil {
			r.render(w, next, level+1)
		}

		if !skip {
			io.WriteString(w, pref+")")
			if level != 0 {
				io.WriteString(w, ",")
			}
		}
	}

	if next := n.NextSibling; next != nil {
		r.render(w, next, level)
	}

	return nil
}

// Option modifies Renderer. You cannot define your own options.
type Option func(*Renderer)

// WithGomponentsAlias tells [Renderer] to use alias for every function call
// from gomponents package in rendered code.
func WithGomponentsAlias(alias string) Option {
	return func(r *Renderer) {
		r.gomponentsPkgName = alias
	}
}

// WithHTMLPackageElements tells [Renderer] to use gomponents/html package in
// order to render standard HTML elements. It uses default gomponent package as
// fallback.
//
// In order to have proper go code rendered, you have to provide
// gomponentsAlias, but you can leave htmlAlias empty.
//
// Empty htmlAlias will render function calls imported into namespace with "."
// import operator.
func WithHTMLPackageElements(gomponentsAlias, htmlAlias string) Option {
	fallback := rawElementsWriter(gomponentsAlias)

	return func(r *Renderer) {
		r.nodeWrite = func(w io.Writer, level int, tag string) error {
			f, ok := htmlElems[strings.ToLower(tag)]
			if !ok {
				return fallback(w, level, tag)
			}

			prefx := "\n"
			if level >= 0 {
				for range level {
					prefx += " "
				}
			}

			pkgName := ""
			if htmlAlias != "" {
				pkgName = fmt.Sprintf("%s.", htmlAlias)
			}

			_, err := fmt.Fprintf(w, "%s%s%s(", prefx, pkgName, f)
			if err != nil {
				return fmt.Errorf("html func tag %s: %w", f, err)
			}

			return nil
		}
	}
}

// WithHTMLPackageElements tells [Renderer] to use gomponents/html package in
// order to render standard HTML attributes. It uses default gomponent package
// as fallback.
//
// In order to have proper go code rendered, you have to provide
// gomponentsAlias, but you can leave htmlAlias empty.
//
// Empty htmlAlias will render function calls imported into namespace with "."
// import operator.
func WithHTMLPackageAttributes(gomponentsAlias, htmlAlias string) Option {
	fallback := rawAttributesWriter(gomponentsAlias)
	return func(r *Renderer) {
		r.attrWrite = func(w io.Writer, level int, a html.Attribute) error {
			prefx := "\n"
			if level >= 0 {
				for range level + 1 {
					prefx += " "
				}
			}

			k := strings.ToLower(a.Key)
			onlyAttr, isOnlyAttr := attrsWithoutValue[k]
			withValue, isWithValue := attrsWithValue[k]

			pkgName := ""
			if htmlAlias != "" {
				pkgName = fmt.Sprintf("%s.", htmlAlias)
			}

			switch {
			case isOnlyAttr:
				_, err := fmt.Fprintf(w, "%s%s%s(),", prefx, pkgName, onlyAttr)
				if err != nil {
					return fmt.Errorf("attr from html package (%s): %w", a.Key, err)
				}
			case isWithValue:
				_, err := fmt.Fprintf(w, "%s%s%s(\"%s\"),", prefx, pkgName, withValue, a.Val)
				if err != nil {
					return fmt.Errorf("attr from html package (%s): %w", a.Key, err)
				}
			case strings.HasPrefix(k, "data-"):
				k = strings.TrimPrefix(k, "data-")
				_, err := fmt.Fprintf(w, "%s%sData(\"%s\", \"%s\"),", prefx, pkgName, k, a.Val)
				if err != nil {
					return fmt.Errorf("attr from html package (%s): %w", a.Key, err)
				}
			case strings.HasPrefix(k, "aria-"):
				k = strings.TrimPrefix(k, "aria-")
				_, err := fmt.Fprintf(w, "%s%sAria(\"%s\", \"%s\"),", prefx, pkgName, k, a.Val)
				if err != nil {
					return fmt.Errorf("attr from html package (%s): %w", a.Key, err)
				}
			default:
				return fallback(w, level, a)
			}

			return nil
		}
	}
}

// WithRenderSVG tells [Renderer] to fully render SVG elements with gomponents
// package.
//
// You can mix this option with [WithHTMLPackageElements] or
// [WithHTMLPackageAttributes] just fine.
//
// By default, [Renderer] will use g.Raw function call to render SVGs.
func WithRenderSVG(r *Renderer) {
	r.renderSVG = true
}

func rawElementsWriter(packageName string) func(io.Writer, int, string) error {
	return func(w io.Writer, level int, tag string) error {
		prefx := "\n"
		if level >= 0 {
			for range level {
				prefx += " "
			}
		}

		_, err := fmt.Fprintf(w, "%s%s.El(\"%s\",", prefx, packageName, tag)
		if err != nil {
			return fmt.Errorf("rawElementsWriter: %w", err)
		}

		return nil
	}
}

func rawAttributesWriter(packageName string) func(io.Writer, int, html.Attribute) error {
	return func(w io.Writer, level int, a html.Attribute) error {
		prefx := "\n"
		if level >= 0 {
			for range level + 1 {
				prefx += " "
			}
		}

		_, err := fmt.Fprintf(w, "%s%s.Attr(\"%s\", \"%s\"),", prefx, packageName, a.Key, a.Val)
		if err != nil {
			return fmt.Errorf("rawAttributesWriter: %w", err)
		}

		return nil
	}
}

var htmlElems = map[string]string{
	"a":          "A",
	"address":    "Address",
	"area":       "Area",
	"article":    "Article",
	"aside":      "Aside",
	"audio":      "Audio",
	"base":       "Base",
	"blockquote": "BlockQuote",
	"body":       "Body",
	"br":         "Br",
	"button":     "Button",
	"canvas":     "Canvas",
	"cite":       "Cite",
	"code":       "Code",
	"col":        "Col",
	"colgroup":   "ColGroup",
	"data":       "DataEl",
	"datalist":   "DataList",
	"details":    "Details",
	"dialog":     "Dialog",
	"div":        "Div",
	"dl":         "Dl",
	"embed":      "Embed",
	"form":       "Form",
	"fieldset":   "FieldSet",
	"figure":     "Figure",
	"footer":     "Footer",
	"head":       "Head",
	"header":     "Header",
	"hgroup":     "HGroup",
	"hr":         "Hr",
	"html":       "HTML",
	"iframe":     "IFrame",
	"img":        "Img",
	"input":      "Input",
	"label":      "Label",
	"legend":     "Legend",
	"li":         "Li",
	"link":       "Link",
	"main":       "Main",
	"menu":       "Menu",
	"meta":       "Meta",
	"meter":      "Meter",
	"nav":        "Nav",
	"noscript":   "NoScript",
	"object":     "Object",
	"ol":         "Ol",
	"optgroup":   "OptGroup",
	"option":     "Option",
	"p":          "P",
	"param":      "Param",
	"picture":    "Picture",
	"pre":        "Pre",
	"progress":   "Progress",
	"script":     "Script",
	"section":    "Section",
	"select":     "Select",
	"source":     "Source",
	"span":       "Span",
	"style":      "StyleEl",
	"summary":    "Summary",
	"svg":        "SVG",
	"table":      "Table",
	"tbody":      "TBody",
	"td":         "Td",
	"textarea":   "Textarea",
	"tfoot":      "TFoot",
	"th":         "Th",
	"thead":      "THead",
	"tr":         "Tr",
	"ul":         "Ul",
	"wbr":        "Wbr",
	"abbr":       "Abbr",
	"b":          "B",
	"caption":    "Caption",
	"dd":         "Dd",
	"del":        "Del",
	"dfn":        "Dfn",
	"dt":         "Dt",
	"em":         "Em",
	"figcaption": "FigCaption",
	"h1":         "H1",
	"h2":         "H2",
	"h3":         "H3",
	"h4":         "H4",
	"h5":         "H5",
	"h6":         "H6",
	"i":          "I",
	"ins":        "Ins",
	"kbd":        "Kbd",
	"mark":       "Mark",
	"q":          "Q",
	"s":          "S",
	"samp":       "Samp",
	"small":      "Small",
	"strong":     "Strong",
	"sub":        "Sub",
	"sup":        "Sup",
	"time":       "Time",
	"title":      "TitleEl",
	"u":          "U",
	"var":        "Var",
	"video":      "Video",
}

var attrsWithoutValue = map[string]string{
	"async":       "Async",
	"autofocus":   "AutoFocus",
	"autoplay":    "AutoPlay",
	"checked":     "Checked",
	"controls":    "Controls",
	"defer":       "Defer",
	"disabled":    "Disabled",
	"loop":        "Loop",
	"multiple":    "Multiple",
	"muted":       "Muted",
	"playsinline": "PlaysInline",
	"readonly":    "ReadOnly",
	"required":    "Required",
	"selected":    "Selected",
}

var attrsWithValue = map[string]string{
	"crossorigin":  "CrossOrigin",
	"datetime":     "DateTime",
	"draggable":    "Draggable",
	"accept":       "Accept",
	"action":       "Action",
	"alt":          "Alt",
	"as":           "As",
	"autocomplete": "AutoComplete",
	"charset":      "Charset",
	"citeattr":     "CiteAttr",
	"class":        "Class",
	"cols":         "Cols",
	"colspan":      "ColSpan",
	"content":      "Content",
	"for":          "For",
	"formattr":     "FormAttr",
	"height":       "Height",
	"href":         "Href",
	"id":           "ID",
	"integrity":    "Integrity",
	"labelattr":    "LabelAttr",
	"lang":         "Lang",
	"list":         "List",
	"loading":      "Loading",
	"max":          "Max",
	"maxlength":    "MaxLength",
	"method":       "Method",
	"min":          "Min",
	"minlength":    "MinLength",
	"name":         "Name",
	"pattern":      "Pattern",
	"placeholder":  "Placeholder",
	"poster":       "Poster",
	"preload":      "Preload",
	"rel":          "Rel",
	"role":         "Role",
	"rows":         "Rows",
	"rowspan":      "RowSpan",
	"src":          "Src",
	"srcset":       "SrcSet",
	"step":         "Step",
	"style":        "Style",
	"tabindex":     "TabIndex",
	"target":       "Target",
	"title":        "Title",
	"type":         "Type",
	"value":        "Value",
	"width":        "Width",
	"enctype":      "EncType",
	"dir":          "Dir",
}

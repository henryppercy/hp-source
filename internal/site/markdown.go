package site

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/frostybee/kazari"
	kazarichroma "github.com/frostybee/kazari/chroma"
	kazarimd "github.com/frostybee/kazari/goldmark"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// imagePattern builds a content image URL from a bare name and target width.
// {path} is the name; {width}, when present, is the requested pixel width.
var imagePattern = "/static/images/{path}"

const (
	coverWidth = 400
	bodyWidth  = 1600
)

// codeTheme is the restrained syntax palette kazari resolves by name: keywords
// in the accent, strings and numbers in one warm tone, comments and punctuation
// greyed, everything else default text. Colours mirror the tokens in input.css.
const codeTheme = "hp-code"

func registerCodeTheme() {
	styles.Register(chroma.MustNewStyle(codeTheme, chroma.StyleEntries{
		chroma.Background:    "#16181a bg:#fafafa",
		chroma.Text:          "#16181a",
		chroma.Comment:       "#54585b",
		chroma.Keyword:       "#b42318",
		chroma.LiteralString: "#9a6a00",
		chroma.LiteralNumber: "#9a6a00",
		chroma.Punctuation:   "#8a8e91",
	}))
}

// newCodeEngine builds the kazari engine that renders fenced code into framed,
// highlighted blocks. It is the single source of truth for the code theme; the
// matching CSS comes from engine.CSS(). The site ships no JavaScript and no dark
// theme, so every block uses the plain code frame in one restrained palette.
func newCodeEngine() *kazari.Engine {
	registerCodeTheme()
	return kazari.New(
		kazari.WithHighlighter(kazarichroma.New()),
		kazari.WithThemes(codeTheme, codeTheme),
		kazari.WithMinContrast(0),
		kazari.WithFrameDetection(false),
		kazari.WithDefaults(kazari.BlockDefaults{Frame: kazari.FrameCode, PreserveIndent: true}),
		kazari.WithCopyButton(false),
		kazari.WithWrapButton(false),
		kazari.WithFullscreenButton(false),
	)
}

func newMarkdown(engine *kazari.Engine) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			figureExt{},
			kazarimd.New(engine),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

// render turns markdown into sanitised-by-construction HTML plus a table of
// contents built from the heading IDs goldmark assigns.
func render(md goldmark.Markdown, source string) (template.HTML, []templates.TOCEntry, error) {
	src := []byte(source)
	doc := md.Parser().Parse(text.NewReader(src))
	resolveImages(doc)
	toc := extractTOC(doc, src)

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, src, doc); err != nil {
		return "", nil, fmt.Errorf("failed to render markdown: %w", err)
	}
	return template.HTML(buf.String()), toc, nil
}

// resolveImages rewrites local image destinations through imageURL, so bodies
// can use bare filenames. Only image nodes are touched (links are left alone),
// as are external (scheme/protocol-relative) and already-based URLs.
func resolveImages(doc ast.Node) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if img, ok := n.(*ast.Image); ok && entering {
			img.Destination = []byte(imageURL(string(img.Destination), bodyWidth))
		}
		return ast.WalkContinue, nil
	})
}

// figureNode wraps an image that is the sole content of a paragraph, so it
// renders as <figure> with its alt text as a visible <figcaption>. Inline images
// are untouched.
type figureNode struct {
	ast.BaseBlock
}

var kindFigure = ast.NewNodeKind("Figure")

func (*figureNode) Kind() ast.NodeKind { return kindFigure }

func (n *figureNode) Dump(source []byte, level int) { ast.DumpHelper(n, source, level, nil, nil) }

// figureExt wires the figure transformer and renderer into goldmark.
type figureExt struct{}

func (figureExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(figureTransformer{}, 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(figureRenderer{}, 500),
	))
}

// figureTransformer replaces each paragraph holding a single image with a figure
// node, lifting the image out of the <p> that would otherwise wrap it.
type figureTransformer struct{}

func (figureTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	var paras []*ast.Paragraph
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if p, ok := n.(*ast.Paragraph); entering && ok && loneImage(p) {
			paras = append(paras, p)
		}
		return ast.WalkContinue, nil
	})
	for _, p := range paras {
		fig := &figureNode{}
		fig.AppendChild(fig, p.FirstChild())
		p.Parent().ReplaceChild(p.Parent(), p, fig)
	}
}

func loneImage(p *ast.Paragraph) bool {
	_, ok := p.FirstChild().(*ast.Image)
	return ok && p.ChildCount() == 1
}

// figureRenderer emits <figure><img><figcaption>alt</figcaption></figure>. The
// inner image renders via goldmark's default (carrying the alt attribute); the
// caption repeats that alt text as visible copy.
type figureRenderer struct{}

func (r figureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindFigure, r.render)
}

func (figureRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<figure>\n")
		return ast.WalkContinue, nil
	}
	if img, ok := n.FirstChild().(*ast.Image); ok {
		if caption := nodeText(img, source); caption != "" {
			w.WriteString("<figcaption>")
			w.Write(util.EscapeHTML([]byte(caption)))
			w.WriteString("</figcaption>\n")
		}
	}
	w.WriteString("</figure>\n")
	return ast.WalkContinue, nil
}

// imageURL resolves a bare image name to its served URL at the given width.
// External URLs (scheme or protocol-relative) are returned untouched.
func imageURL(name string, width int) string {
	if name == "" || hasScheme(name) || strings.HasPrefix(name, "//") {
		return name
	}
	url := strings.ReplaceAll(imagePattern, "{path}", strings.TrimPrefix(name, "/"))
	return strings.ReplaceAll(url, "{width}", strconv.Itoa(width))
}

// hasScheme reports whether dest carries a URL scheme (http:, data:, ...),
// i.e. a colon before any path separator.
func hasScheme(dest string) bool {
	i := strings.Index(dest, ":")
	return i > 0 && !strings.ContainsAny(dest[:i], "/?#")
}

// extractTOC collects depth-2 headings, nesting depth-3 headings under the
// preceding depth-2 entry.
func extractTOC(doc ast.Node, source []byte) []templates.TOCEntry {
	var toc []templates.TOCEntry
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok || (h.Level != 2 && h.Level != 3) {
			return ast.WalkContinue, nil
		}

		entry := templates.TOCEntry{Title: nodeText(h, source), Anchor: headingID(h)}
		if h.Level == 3 && len(toc) > 0 {
			parent := &toc[len(toc)-1]
			parent.Children = append(parent.Children, entry)
		} else {
			toc = append(toc, entry)
		}
		return ast.WalkSkipChildren, nil
	})
	return toc
}

func headingID(n ast.Node) string {
	if v, ok := n.AttributeString("id"); ok {
		switch id := v.(type) {
		case []byte:
			return string(id)
		case string:
			return id
		}
	}
	return ""
}

func nodeText(n ast.Node, source []byte) string {
	var b strings.Builder
	ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := c.(*ast.Text); ok {
				b.Write(t.Segment.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

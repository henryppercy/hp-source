package site

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/frostybee/kazari"
	kazarichroma "github.com/frostybee/kazari/chroma"
	kazarimd "github.com/frostybee/kazari/goldmark"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// imageBase is where local images are served from. Content references bare
// filenames (glasto.jpg); resolveImages and coverURL prepend this. Swap it for
// a CDN host to relocate every image without touching content.
const imageBase = "/static/images"

// newCodeEngine builds the kazari engine that renders fenced code into framed,
// highlighted blocks. It is the single source of truth for the code theme and
// which controls exist; the matching CSS/JS come from engine.CSS()/JS(). Themes
// beyond the active light/dark pair are mapped so per-block theme="..." resolves.
func newCodeEngine() *kazari.Engine {
	hl := kazarichroma.New(kazarichroma.WithStyleMap(map[string]string{
		"github":      "github",
		"github-dark": "github-dark",
		"dracula":     "dracula",
		"monokai":     "monokai",
		"onedark":     "onedark",
		"nord":        "nord",
	}))
	return kazari.New(
		kazari.WithHighlighter(hl),
		kazari.WithThemes("github", "github-dark"),
		kazari.WithDarkMode(kazari.MediaQueryMode()),
		kazari.WithCopyButton(true),
		kazari.WithWrapButton(true),
		kazari.WithFullscreenButton(true),
		kazari.WithCollapsible(kazari.CollapsibleConfig{LineThreshold: 30, PreviewLines: 3}),
	)
}

func newMarkdown(engine *kazari.Engine) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			kazarimd.New(engine),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
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

// resolveImages rewrites local image destinations under imageBase, so bodies
// can use bare filenames. Only image nodes are touched (links are left alone),
// as are external (scheme/protocol-relative) and already-based URLs.
func resolveImages(doc ast.Node) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if img, ok := n.(*ast.Image); ok && entering {
			img.Destination = []byte(imageURL(string(img.Destination)))
		}
		return ast.WalkContinue, nil
	})
}

func imageURL(dest string) string {
	if dest == "" || hasScheme(dest) ||
		strings.HasPrefix(dest, "//") || strings.HasPrefix(dest, imageBase+"/") {
		return dest
	}
	return imageBase + "/" + strings.TrimPrefix(dest, "/")
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

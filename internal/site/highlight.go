// highlight.go

package site

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// markNode wraps text delimited by "==...==", rendered as <mark>. It sits
// alongside strong/em rather than replacing either, so the two can combine.
type markNode struct {
	ast.BaseInline
}

var kindMark = ast.NewNodeKind("Highlight")

func (*markNode) Kind() ast.NodeKind { return kindMark }

func (n *markNode) Dump(source []byte, level int) { ast.DumpHelper(n, source, level, nil, nil) }

func newMarkNode() *markNode { return &markNode{} }

type markDelimiterProcessor struct{}

func (p *markDelimiterProcessor) IsDelimiter(b byte) bool { return b == '=' }

func (p *markDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *markDelimiterProcessor) OnMatch(consumes int) ast.Node { return newMarkNode() }

var defaultMarkDelimiterProcessor = &markDelimiterProcessor{}

type markParser struct{}

var defaultMarkParser = &markParser{}

// NewMarkParser returns a new InlineParser that parses "==highlight==".
func NewMarkParser() parser.InlineParser { return defaultMarkParser }

func (s *markParser) Trigger() []byte { return []byte{'='} }

func (s *markParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 2, defaultMarkDelimiterProcessor)
	if node == nil || node.OriginalLength > 2 || before == '=' {
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *markParser) CloseBlock(parent ast.Node, pc parser.Context) {
	// nothing to do
}

// MarkHTMLRenderer renders markNode as <mark>.
type MarkHTMLRenderer struct{}

func NewMarkHTMLRenderer() renderer.NodeRenderer { return &MarkHTMLRenderer{} }

func (r *MarkHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindMark, r.renderMark)
}

func (r *MarkHTMLRenderer) renderMark(
	w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<mark>")
	} else {
		w.WriteString("</mark>")
	}
	return ast.WalkContinue, nil
}

// markExt is an extension that lets you write highlighted text as "==text==".
type markExt struct{}

// Mark is the extension instance; add it via goldmark.WithExtensions(Mark).
var Mark = &markExt{}

func (e *markExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewMarkParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewMarkHTMLRenderer(), 500),
	))
}

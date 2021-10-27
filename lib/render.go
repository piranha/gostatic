// (c) 2012 Alexander Solovyov
// under terms of ISC license

package gostatic

import (
	"bytes"
	"fmt"
	"strings"

	chroma "github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func Markdown(source string) string {
	
	md := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chroma.WithLineNumbers(false),
					chroma.WithPreWrapper(&preWrapStruct{}),
				),
			),
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		errhandle(err)
		return ""
	}

	return buf.String()
}

type preWrapStruct struct {
	wroteScript bool
}

const script = `
<script>
function getId(btn)
{
	navigator.clipboard.writeText(btn.nextElementSibling.innerText).catch((error) => {
		console.error(error);
	})
}
</script>
`

const start = `
<button class="copy-code-button" onclick="getId(this)" type="button">Copy</button>
<pre class="highlight"  tabindex="0" %s><code>`

func (p *preWrapStruct) Start(code bool, styleAttr string) string {
	w := &strings.Builder{}
	if !p.wroteScript {
		p.wroteScript = true
		fmt.Fprint(w, script)
	}

	fmt.Fprintf(w, start, styleAttr)
	return w.String()
}

func (p *preWrapStruct) End(code bool) string {
	return "</code></pre>"
}

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

func Markdown(source string, args []string) string {

	extensions := []goldmark.Extender{
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		extension.GFM,
		extension.DefinitionList,
		extension.Footnote,
		extension.Typographer,
	}

	for _, v := range args {
		//chroma=monokai is a code highlighting style example
		if strings.HasPrefix(v, "chroma=") {
			style := strings.Replace(v, "chroma=", "", 1)

			extensions = append(extensions, highlighting.NewHighlighting(
				highlighting.WithStyle(style),
				highlighting.WithFormatOptions(
					chroma.WithLineNumbers(false),
					chroma.WithPreWrapper(&preWrapStruct{}),
				),
			))
		}
	}

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
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
}

const start = `
<div class="highlight">
<pre %s>
<code>`

func (p *preWrapStruct) Start(code bool, styleAttr string) string {
	w := &strings.Builder{}

	if strings.HasSuffix(styleAttr, `"`) {
		fmt.Fprintf(w, start, styleAttr[:len(styleAttr)-1])
	} else {
		fmt.Fprintf(w, start, `style="`)
	}

	return w.String()
}

func (p *preWrapStruct) End(code bool) string {
	return `</code></pre></div>`
}

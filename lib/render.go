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
				highlighting.WithGuessLanguage(true), // this makes sure lines without language dont look bad! re:(^```$)
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

const start = `<pre %s><code>`

func (p *preWrapStruct) Start(code bool, styleAttr string) string {
	w := &strings.Builder{}

	styleAttr = strings.TrimSpace(styleAttr) // this param has spaces sometimes

	if strings.HasPrefix(styleAttr, `style="`) {
		newStyle := styleAttr[:len(styleAttr)-1] + `;overflow-x: auto"`
		fmt.Fprintf(w, start, newStyle)
	} else {
		// styleAttr doesn't start with 'style='
		fmt.Fprintf(w, start, `style="overflow-x: auto"`)
	}

	return w.String()
}

func (p *preWrapStruct) End(code bool) string {
	return `</code></pre>`
}

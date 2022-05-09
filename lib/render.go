// (c) 2012 Alexander Solovyov
// under terms of ISC license

package gostatic

import (
	"bytes"
	"fmt"
	"strings"
	"regexp"
	"os"
	"html"

	chroma "github.com/alecthomas/chroma"
	chromahtml "github.com/alecthomas/chroma/formatters/html"
	chromalexers "github.com/alecthomas/chroma/lexers"
	chromastyles "github.com/alecthomas/chroma/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	markhtml "github.com/yuin/goldmark/renderer/html"
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
					chromahtml.WithLineNumbers(false),
					chromahtml.WithPreWrapper(&preWrapStruct{}),
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
			markhtml.WithXHTML(),
			markhtml.WithUnsafe(),
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
		newStyle := styleAttr[:len(styleAttr)-1] + `; overflow-x: auto"`
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


var Code = regexp.MustCompile(`(?s)<pre><code[^>]*>.+?</code></pre>`)
var LangRe = regexp.MustCompile(`<code class="language-([^"]+)">`)

func Chroma(htmlsource string, style string) string {
	f := chromahtml.New(
		chromahtml.WithLineNumbers(false),
		chromahtml.WithPreWrapper(&preWrapStruct{}),
	)

	return Code.ReplaceAllStringFunc(htmlsource, func (source string) string {
		pre_start := strings.IndexByte(source, '>') + 1
		code_start := strings.IndexByte(source[pre_start:], '>') + pre_start + 1
		code_end := len(source) - 13 // 13 is len('</code></pre>')
		code := html.UnescapeString(source[code_start:code_end])

		// get lexer
		m := LangRe.FindStringSubmatch(source)
		var lexer chroma.Lexer
		if len(m) > 1 {
			lexer = chromalexers.Get(m[1])
		}
		if lexer == nil {
			lexer = chromalexers.Analyse(code)
		}
		if lexer == nil {
			lexer = chromalexers.Fallback
		}

		// get style
		s := chromastyles.Get(style)
		if s == nil {
			s = chromastyles.Fallback
		}

		// tokenize
		it, err := lexer.Tokenise(nil, code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Exception: %v\n", err)
			return source
		}

		var b bytes.Buffer
		f.Format(&b, s, it)
		return b.String()
	})
}

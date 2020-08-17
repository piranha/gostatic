package processors

import (
	"mime"
	"path/filepath"
	"regexp"

	gostatic "github.com/piranha/gostatic/lib"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

type MinifyProcessor struct {
}

func NewMinifyProcessor() *MinifyProcessor {
	return &MinifyProcessor{}
}

func (p *MinifyProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessPage(page, args)
}

func (p *MinifyProcessor) Description() string {
	return "minify supported file types"
}

func (p *MinifyProcessor) Mode() int {
	return gostatic.Post
}

func ProcessPage(page *gostatic.Page, args []string) error {
	m := minify.New()

	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	fileMime := mime.TypeByExtension(filepath.Ext(page.OutputPath()))

	s, err := m.String(fileMime, page.Content())
	if err != nil {
		return err
	}

	page.SetContent(s)

	return nil
}

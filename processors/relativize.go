package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
	"regexp"
)

type RelativizeProcessor struct {
}

func NewRelativizeProcessor() *RelativizeProcessor {
	return &RelativizeProcessor{}
}

func (p *RelativizeProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessRelativize(page, args)
}

func (p *RelativizeProcessor) Description() string {
	return "make all urls bound at root relative " +
		"(allows deploying resulting site in a subdirectory)"
}

func (p *RelativizeProcessor) Mode() int {
	return 0
}

var RelRe = regexp.MustCompile(`(href|src)=["']/([^"']*)["']`)
var NonProtoRe = regexp.MustCompile(`(href|src)=["']//`)

func ProcessRelativize(page *gostatic.Page, args []string) error {
	repl := `$1="` + page.Rel("/") + `$2"`
	rv := RelRe.ReplaceAllStringFunc(page.Content(), func(inp string) string {
		if NonProtoRe.MatchString(inp) {
			return inp
		}
		return RelRe.ReplaceAllString(inp, repl)
	})
	page.SetContent(rv)
	return nil
}

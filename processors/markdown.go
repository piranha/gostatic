package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
	bf "github.com/russross/blackfriday"
)

type MarkdownProcessor struct {
}

func NewMarkdownProcessor() *MarkdownProcessor {
	return &MarkdownProcessor{}
}

func (p *MarkdownProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessMarkdown(page, args)
}

func (p *MarkdownProcessor) Description() string {
	return "process content as a markdown"
}

func (p *MarkdownProcessor) Mode() int {
	return 0
}

func ProcessMarkdown(page *gostatic.Page, args []string) error {
	result := Markdown(page.Content())
	page.SetContent(result)
	return nil
}

func Markdown(source string) string {
	// set up the HTML renderer
	flags := 0
	flags |= bf.HTML_USE_SMARTYPANTS
	flags |= bf.HTML_SMARTYPANTS_FRACTIONS
	renderer := bf.HtmlRenderer(flags, "", "")

	// set up the parser
	ext := 0
	ext |= bf.EXTENSION_NO_INTRA_EMPHASIS
	ext |= bf.EXTENSION_TABLES
	ext |= bf.EXTENSION_FENCED_CODE
	ext |= bf.EXTENSION_AUTOLINK
	ext |= bf.EXTENSION_STRIKETHROUGH
	ext |= bf.EXTENSION_SPACE_HEADERS
	ext |= bf.EXTENSION_FOOTNOTES
	ext |= bf.EXTENSION_HEADER_IDS

	return string(bf.Markdown([]byte(source), renderer, ext))
}

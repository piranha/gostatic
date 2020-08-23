package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
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
	result := gostatic.Markdown(page.Content())
	page.SetContent(result)
	return nil
}

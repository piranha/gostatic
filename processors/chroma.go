package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
)

type ChromaProcessor struct {
}

func NewChromaProcessor() *ChromaProcessor {
	return &ChromaProcessor{}
}

func (p *ChromaProcessor) Process(page *gostatic.Page, args []string) error {
	style := "monokai"
	if len(args) > 0 {
		style = args[0]
	}
	result := gostatic.Chroma(page.Content(), style)
	page.SetContent(result)
	return nil
}

func (p *ChromaProcessor) Description() string {
	return "process content as a markdown"
}

func (p *ChromaProcessor) Mode() int {
	return 0
}

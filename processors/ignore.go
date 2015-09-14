package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
)

type IgnoreProcessor struct {
}

func NewIgnoreProcessor() *IgnoreProcessor {
	return &IgnoreProcessor{}
}

func (p *IgnoreProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessIgnore(page, args)
}

func (p *IgnoreProcessor) Description() string {
	return "ignore file"
}

func (p *IgnoreProcessor) Mode() int {
	return gostatic.Pre
}

func ProcessIgnore(page *gostatic.Page, args []string) error {
	page.SetState(gostatic.StateIgnored)
	return nil
}

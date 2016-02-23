package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
	"time"
)

type IgnoreFutureProcessor struct {
}

func NewIgnoreFutureProcessor() *IgnoreFutureProcessor {
	return &IgnoreFutureProcessor{}
}

func (p *IgnoreFutureProcessor) Process(page *gostatic.Page, args []string) error {
	if time.Now().Before(page.Date) {
		return ProcessIgnore(page, args)
	}
	return nil
}

func (p *IgnoreFutureProcessor) Description() string {
	return "ignore file dated in future"
}

func (p *IgnoreFutureProcessor) Mode() int {
	return gostatic.Pre
}

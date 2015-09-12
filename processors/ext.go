package processors

import (
	"errors"
	gostatic "github.com/piranha/gostatic/lib"
	"path/filepath"
)

type ExtProcessor struct {
}

func NewExtProcessor() *ExtProcessor {
	return &ExtProcessor{}
}

func (p *ExtProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessExt(page, args)
}

func (p *ExtProcessor) Description() string {
	return "change extension"
}

func (p *ExtProcessor) Mode() int {
	return gostatic.Pre
}

func ProcessExt(page *gostatic.Page, args []string) error {
	if len(args) < 1 {
		return errors.New(
			"'ext' rule requires an extension prefixed with dot")
	}
	newExt := args[0]

	ext := filepath.Ext(page.Path)
	if ext == "" {
		page.Path = page.Path + newExt
	} else {
		page.Path = page.Path[0:len(page.Path)-len(ext)] + newExt
	}
	return nil
}

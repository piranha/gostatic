package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
	"path/filepath"
	"strings"
)

type DirectorifyProcessor struct {
}

func NewDirectorifyProcessor() *DirectorifyProcessor {
	return &DirectorifyProcessor{}
}

func (p *DirectorifyProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessDirectorify(page, args)
}

func (p *DirectorifyProcessor) Description() string {
	return "path/name.html -> path/name/index.html"
}

func (p *DirectorifyProcessor) Mode() int {
	return gostatic.Pre
}

func ProcessDirectorify(page *gostatic.Page, args []string) error {
	if filepath.Base(page.Path) != "index.html" {
		page.Path = strings.Replace(page.Path, ".html", "/index.html", 1)
	}
	return nil
}

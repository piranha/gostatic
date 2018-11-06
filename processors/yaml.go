package processors

import (
	"regexp"

	gostatic "github.com/piranha/gostatic/lib"
)

type YamlProcessor struct {
}

func NewYamlProcessor() *YamlProcessor {
	return &YamlProcessor{}
}

func (p *YamlProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessYaml(page, args)
}

func (p *YamlProcessor) Description() string {
	return "read config from content using yaml format"
}

func (p *YamlProcessor) Mode() int {
	return gostatic.Pre
}

func ProcessYaml(page *gostatic.Page, args []string) error {
	r, err := regexp.Compile(`(?sm)(^---\r?\n.*)^---\r?\n(.*)`)
	if err != nil {
		panic(err)
	}
	parts := r.FindStringSubmatch(page.Content())

	if len(parts) != 3 {
		// no configuration, well then...
		page.PageHeader = *gostatic.NewPageHeader()
		return nil
	}

	page.PageHeader = *gostatic.ParseYamlHeader(parts[1])
	page.SetContent(parts[2])
	return nil
}

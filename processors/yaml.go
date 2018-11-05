package processors

import (
	"regexp"

	gostatic "github.com/piranha/gostatic/lib"
)

// YamlProcessor is empty struct
type YamlProcessor struct {
}

// NewYamlProcessor is constructor for YamlProcessor
func NewYamlProcessor() *YamlProcessor {
	return &YamlProcessor{}
}

// Process is function for runnig ProcessYaml
func (p *YamlProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessYaml(page, args)
}

// Description is function for write description
func (p *YamlProcessor) Description() string {
	return "read config from content using yaml format"
}

// Mode is work mode for this processor
func (p *YamlProcessor) Mode() int {
	return gostatic.Pre
}

// ProcessYaml is function for proccess yaml header
func ProcessYaml(page *gostatic.Page, args []string) error {
	r, err := regexp.Compile(`(?sm)^---\r?\n(.*)^---\r?\n(.*)`)
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

package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
)

type ConfigProcessor struct {
}

func NewConfigProcessor() *ConfigProcessor {
	return &ConfigProcessor{}
}

func (p *ConfigProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessConfig(page, args)
}

func (p *ConfigProcessor) Description() string {
	return "read config from content (separated by '----\\n')"
}

func (p *ConfigProcessor) Mode() int {
	return gostatic.Pre
}

func ProcessConfig(page *gostatic.Page, args []string) error {
	parts := gostatic.TrimSplitN(page.Content(), "\n----\n", 2)
	if len(parts) != 2 {
		// no configuration, well then...
		page.PageHeader = *gostatic.NewPageHeader()
		return nil
	}

	page.PageHeader = *gostatic.ParseHeader(parts[0])
	page.SetContent(parts[1])
	return nil
}

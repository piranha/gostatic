package processors

import (
	"regexp"

	gostatic "github.com/piranha/gostatic/lib"
)

var shortSeparator = regexp.MustCompile(`(?m:^----?\r?\n)`)
var oldSeparator = regexp.MustCompile(`(?m:^----\r?\n)`)

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
	loc := shortSeparator.FindStringIndex(page.Content())

	if (loc != nil) && (loc[0] == 0) {
		// this branch parses Github-style frontmatter, i.e.
		//
		//     ---
		//     date: 1234-12-12
		//     ---
		parts := shortSeparator.Split(page.Content(), 3)
		// page starts with a separator but then no second separator? This means
		// no configuration is present.
		if len(parts) != 3 {
			page.PageHeader = *gostatic.NewPageHeader()
			return nil
		}
		page.PageHeader = *gostatic.ParseHeader(parts[1])
		page.SetContent(parts[2])
	} else {
		// this branch parses old gostatic-style frontmatter, i.e.
        //
		//     date: 1234-12-12
		//     ----
		parts := oldSeparator.Split(page.Content(), 2)
		if len(parts) != 2 {
			// no separator to split content? No configuration is present then.
			page.PageHeader = *gostatic.NewPageHeader()
			return nil
		}
		page.PageHeader = *gostatic.ParseHeader(parts[0])
		page.SetContent(parts[1])
	}
	return nil
}

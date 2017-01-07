package processors

import (
	"errors"
	"fmt"
	gostatic "github.com/piranha/gostatic/lib"
	"strings"
	"time"
)

type TagsProcessor struct {
}

func NewTagsProcessor() *TagsProcessor {
	return &TagsProcessor{}
}

func (p *TagsProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessTags(page, args)
}

func (p *TagsProcessor) Description() string {
	return "generate tags pages for tags mentioned in page header " +
		"(argument - tag template)"
}

func (p *TagsProcessor) Mode() int {
	return gostatic.Pre
}

func ProcessTags(page *gostatic.Page, args []string) error {
	if len(args) < 1 {
		return errors.New("'tags' rule needs an argument")
	}
	pathPattern := args[0]

	if page.Tags == nil {
		return nil
	}

	site := page.Site

	for _, tag := range page.Tags {
		tagpath := strings.Replace(pathPattern, "*", tag, 1)

		if site.Pages.BySource(tagpath) == nil {
			pattern, rules := site.Rules.MatchedRules(tagpath)
			if rules == nil {
				return fmt.Errorf("Tag path '%s' does not match any rule", tagpath)
			}
			if len(rules) > 1 {
				return fmt.Errorf("Tags are not supported with multiple rules. Tag in question: '%s'", tagpath)
			}

			tagpage := &gostatic.Page{
				PageHeader: gostatic.PageHeader{Title: tag},
				Site:       site,
				Pattern:    pattern,
				Rule:       rules[0],
				Source:     tagpath,
				Path:       tagpath,
				// tags are never new, because they only depend on pages and
				// have not a bit of original content
				ModTime: time.Unix(0, 0),
			}
			tagpage.SetWasRead(true)
			page.Site.Pages = append(page.Site.Pages, tagpage)
			tagpage.Peek()
		}
	}

	return nil
}

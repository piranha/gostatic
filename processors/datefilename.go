package processors

import (
	"regexp"
	"time"
	"path/filepath"

	gostatic "github.com/piranha/gostatic/lib"
)

type DatefilenameProcessor struct {
}

func NewDatefilenameProcessor() *DatefilenameProcessor {
	return &DatefilenameProcessor{}
}

func (p *DatefilenameProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessDatefilename(page, args)
}

func (p *DatefilenameProcessor) Description() string {
	return "process filename 2014-05-06-name.md to path /name.html and set page.Date to 2014-05-06"
}

func (p *DatefilenameProcessor) Mode() int {
	return gostatic.Pre
}

func ProcessDatefilename(page *gostatic.Page, args []string) error {
	name := page.Name()
	dir := filepath.Dir(page.Path)

	validName := regexp.MustCompile(`(?P<Year>\d{4})-(?P<Month>\d{2})-(?P<Day>\d{2})-(.*)`)
	if validName.MatchString(name) {
		date := validName.FindStringSubmatch(name)
		page.Path = date[4]
		value := date[1] + "-" + date[2] + "-" + date[3]
		t, err := time.Parse("2006-01-02", value)
		if err == nil {
			page.Date = t
			page.Path = dir + "/" + date[4]
		}
	}
	return nil
}

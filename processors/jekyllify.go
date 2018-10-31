package processors

import (
	"os"
	"regexp"

	gostatic "github.com/piranha/gostatic/lib"
)

// JekyllifyProcessor is new structure
type JekyllifyProcessor struct {
}

// NewJekyllifyProcessor is constructor for JekyllifyProcessor
func NewJekyllifyProcessor() *JekyllifyProcessor {
	return &JekyllifyProcessor{}
}

// Process is function for executing ProcessJekyllify function
func (p *JekyllifyProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessJekyllify(page, args)
}

// Description is function for writing description
func (p *JekyllifyProcessor) Description() string {
	return "process filename 2014-05-06-name.md to path /2014/05/06/name.html as pretty permalink on Jekyll"
}

// Mode is work mode for this processor
func (p *JekyllifyProcessor) Mode() int {
	return gostatic.Pre
}

// ProcessJekyllify is function for processing pages
func ProcessJekyllify(page *gostatic.Page, args []string) error {
	name := page.Name()

	validName := regexp.MustCompile(`(?P<Year>\d{4})-(?P<Month>\d{2})-(?P<Day>\d{2})-(.*)`)
	if validName.MatchString(name) {
		date := validName.FindStringSubmatch(name)
		page.Path = date[1] + string(os.PathSeparator) + date[2] + string(os.PathSeparator) + date[3] + string(os.PathSeparator) + date[4]
	}
	return nil
}

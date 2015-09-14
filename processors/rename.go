package processors

import (
	"fmt"
	gostatic "github.com/piranha/gostatic/lib"
	"path/filepath"
	"regexp"
	"strings"
)

type RenameProcessor struct {
}

func NewRenameProcessor() *RenameProcessor {
	return &RenameProcessor{}
}

func (p *RenameProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessRename(page, args)
}

func (p *RenameProcessor) Description() string {
	return "rename resulting file (argument - pattern for renaming, " +
		"relative to current file location)"
}

func (p *RenameProcessor) Mode() int {
	return gostatic.Pre
}
func ProcessRename(page *gostatic.Page, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("'rename' rule needs an argument")
	}
	dest := args[0]

	if strings.Contains(dest, "*") {
		if !strings.Contains(page.Pattern, "*") {
			return fmt.Errorf(
				"'rename' rule cannot rename '%s' to '%s'",
				page.Pattern, dest)
		}

		group := fmt.Sprintf("([^%c]*)", filepath.Separator)
		base := filepath.Base(page.Pattern)
		pat := strings.Replace(regexp.QuoteMeta(base), "\\*", group, 1)

		re, err := regexp.Compile(pat)
		if err != nil {
			return err
		}
		m := re.FindStringSubmatch(filepath.Base(page.Path))

		dest = strings.Replace(dest, "*", m[1], 1)
	}

	page.Path = filepath.Join(filepath.Dir(page.Path), dest)
	return nil
}

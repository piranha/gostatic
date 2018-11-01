package processors

import (
	"bytes"
	"errors"
	"fmt"

	gostatic "github.com/piranha/gostatic/lib"
)

type TemplateProcessor struct {
	inner bool
}

func NewTemplateProcessor() *TemplateProcessor {
	return &TemplateProcessor{}
}

func NewInnerTemplateProcessor() *TemplateProcessor {
	return &TemplateProcessor{true}
}

func (p *TemplateProcessor) Process(page *gostatic.Page, args []string) error {
	if p.inner {
		return ProcessInnerTemplate(page, args)
	}
	return ProcessTemplate(page, args)
}

func (p *TemplateProcessor) Description() string {
	return "put content in a template (argument - template name)"
}

func (p *TemplateProcessor) Mode() int {
	return 0
}

func ProcessTemplate(page *gostatic.Page, args []string) error {
	if len(args) < 1 {
		return errors.New("'template' rule needs an argument")
	}
	pagetype := args[0]
	//todo catch thiss
	defer func() {
		if err := recover(); err != nil {
			//return err //errors.New(fmt.Sprintf("%s: %s", page.Source, err))
		}
	}()

	var buffer bytes.Buffer
	err := page.Site.Template.ExecuteTemplate(&buffer, pagetype, page)
	if err != nil {
		return errors.New(fmt.Sprintf("%s: %s", page.Source, err))
	}

	page.SetContent(buffer.String())
	return nil
}

func ProcessInnerTemplate(page *gostatic.Page, args []string) error {
	//todo catch
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s: %s\n", page.Source, err)
			//return fmt.Sprintf("%s: %s", page.Source, err)
		}
	}()

	t, err := page.Site.Template.Clone()
	if err != nil {
		return err
	}
	t, err = t.New("ad-hoc").Parse(page.Content())
	if err != nil {
		return fmt.Errorf("Page %s: %s", page.Source, err)
	}

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, "ad-hoc", page)
	if err != nil {
		return fmt.Errorf("Page %s: %s", page.Source, err)
	}

	page.SetContent(buffer.String())
	return nil
}

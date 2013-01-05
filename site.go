// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Site struct {
	SiteConfig
	Template *template.Template
	Pages    PageSlice
}

func NewSite(config *SiteConfig) *Site {
	template, err := template.ParseFiles(config.Templates...)
	errhandle(err)
	template.Funcs(funcMap)

	site := &Site{
		SiteConfig: *config,
		Template:   template,
		Pages:      make(PageSlice, 0),
	}

	site.Collect()
	site.FindDeps()

	return site
}

func (site *Site) AddPage(path string) {
	page := NewPage(site, path)
	if page.state != StateIgnored {
		site.Pages = append(site.Pages, page)
	}
}

func (site *Site) Collect() {
	errors := make(chan error)

	filepath.Walk(site.Source, site.collectFunc(errors))

	select {
	case err := <-errors:
		errhandle(err)
	default:
	}

	site.Pages.Sort()
}

func (site *Site) collectFunc(errors chan<- error) filepath.WalkFunc {
	return func(fn string, fi os.FileInfo, err error) error {
		if err != nil {
			errors <- err
			return nil
		}

		if !fi.IsDir() && !strings.HasPrefix(filepath.Base(fn), ".") {
			site.AddPage(fn)
		}

		return nil
	}
}

func (site *Site) FindDeps() {
	for _, page := range site.Pages {
		page.findDeps()
	}
}

func (site *Site) Process() int {
	processed := 0
	for _, page := range site.Pages {
		if page.Changed() {
			page.Process()
			processed++
		}
	}
	return processed
}

func (site *Site) ProcessAll() {
	for _, page := range site.Pages {
		page.Process()
	}
}

func (site *Site) Summary() {
	site.ProcessAll()
	out("Total pages to render: %d\n", len(site.Pages))

	for _, page := range site.Pages {
		out("%s - %s: %d chars; %s\n",
			page.Path, page.Title, len(page.Content()), page.Rule)
		out("------------")
		_, err := page.WriteTo(os.Stdout)
		errhandle(err)
		out("------------\n")
	}
}

func (site *Site) Render() {
	processed := site.Process()
	out("Total pages to render: %d\n", processed)

	for _, page := range site.Pages {
		if !page.Changed() {
			continue
		}

		debug("Rendering %s...\n", page.OutputPath())

		err := os.MkdirAll(filepath.Dir(page.OutputPath()), 0755)
		errhandle(err)

		file, err := os.Create(page.OutputPath())
		errhandle(err)
		page.WriteTo(file)
		file.Close()
	}
}

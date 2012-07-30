// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type Site struct {
	Path     string
	Output   string
	Template *template.Template
	Rules    RuleMap
	Pages    PageSlice
}

func NewSite(config *GlobalConfig) *Site {
	template, err := template.ParseFiles(config.Templates...)
	errhandle(err)

	site := &Site{
		Path:     config.Source,
		Output:   config.Output,
		Template: template,
		Rules:    config.Rules,
		Pages:    make(PageSlice, 0),
	}

	site.Collect()

	return site
}

func (site *Site) AddPage(path string) {
	page := NewPage(site, path)
	site.Pages = append(site.Pages, page)
}

func (site *Site) Collect() {
	errors := make(chan error)

	filepath.Walk(site.Path, site.walkFunc(errors))

	select {
	case err := <-errors:
		errhandle(err)
	default:
	}
}

func (site *Site) walkFunc(errors chan<- error) filepath.WalkFunc {
	return func(fn string, fi os.FileInfo, err error) error {
		if err != nil {
			errors <- err
			return nil
		}

		if !fi.IsDir() {
			site.AddPage(fn)
		}

		return nil
	}
}

func (site *Site) Summary() {
	println("Total pages", len(site.Pages))
	for _, page := range site.Pages {
		page.Process()
	}

	for _, page := range site.Pages {
		fmt.Printf("%s - %s: %d chars; %s\n",
			page.Path, page.Title, len(page.Content), page.Rules)
		fmt.Println("------------")
		_, err := page.WriteTo(os.Stdout)
		errhandle(err)
		fmt.Println("------------\n")
	}
}

func (site *Site) Render() {
	pages := make(PageSlice, len(site.Pages))
	copy(pages, site.Pages)
	for _, page := range pages {
		page.Process()
	}

	// we are doing a second go here because certain processors can modify Pages
	// list
	fmt.Printf("Total pages: %d\n", len(site.Pages))

	for _, page := range site.Pages {
		path := filepath.Join(site.Output, page.Path)

		err := os.MkdirAll(filepath.Dir(path), 0755)
		errhandle(err)

		file, err := os.Create(path)
		errhandle(err)
		page.WriteTo(file)
		file.Close()
	}
}

// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"fmt"
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

	return site
}

func (site *Site) AddPage(path string) {
	page := NewPage(site, path)
	site.Pages = append(site.Pages, page)
}

func (site *Site) Collect() {
	errors := make(chan error)

	filepath.Walk(site.Source, site.walkFunc(errors))

	select {
	case err := <-errors:
		errhandle(err)
	default:
	}

	site.Pages.Sort()
}

func (site *Site) walkFunc(errors chan<- error) filepath.WalkFunc {
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

func (site *Site) Process() {
	for _, page := range site.Pages {
		page.FindDeps()
	}

	for _, page := range site.Pages {
		page.Process()
	}
}

func (site *Site) Summary() {
	site.Process()
	fmt.Printf("Total pages rendered: %d\n", len(site.Pages))

	for _, page := range site.Pages {
		fmt.Printf("%s - %s: %d chars; %s\n",
			page.Path, page.Title, len(page.Content), page.Rule)
		fmt.Println("------------")
		_, err := page.WriteTo(os.Stdout)
		errhandle(err)
		fmt.Println("------------\n")
	}
}

func (site *Site) Render() {
	site.Process()
	fmt.Printf("Total pages rendered: %d\n", len(site.Pages))

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

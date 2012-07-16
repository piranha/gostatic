package main

import (
	"os"
	"fmt"
	"path/filepath"
	"text/template"
)

type Site struct {
	Base string
	Template *template.Template
	Pages []*Page
}


func NewSite(t *template.Template, dir string) *Site {
	site := &Site{dir, t, *new([]*Page)}

	site.Collect()

	return site
}

func (site *Site) AddPage(path string) {
	page := NewPage(path, site.Base, site.Template)
	site.Pages = append(site.Pages, page)
}

func (site *Site) Collect() {
	errors := make(chan error)

    filepath.Walk(site.Base, site.walkFunc(errors))

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
		fmt.Printf("%s - %s: %d chars\n",
			page.Path, page.Title, len(page.Content))
	}
}

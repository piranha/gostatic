// (c) 2012 Alexander Solovyov
// under terms of ISC license

package gostatic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Site struct {
	SiteConfig
	Template  *template.Template
	ChangedAt time.Time
	Pages     PageSlice

	ForceRefresh bool

	mx sync.Mutex

	Processors map[string]Processor
}

func NewSite(config *SiteConfig, procs ProcessorMap) *Site {
	template := template.New("no-idea-what-to-pass-here").Funcs(TemplateFuncMap)
	template, err := template.ParseFiles(config.Templates...)
	errhandle(err)

	changed := config.changedAt
	for _, fn := range config.Templates {
		stat, err := os.Stat(fn)
		if err == nil && changed.Before(stat.ModTime()) {
			changed = stat.ModTime()
		}
	}

	site := &Site{
		SiteConfig: *config,
		Template:   template,
		ChangedAt:  changed,
		Pages:      make(PageSlice, 0),
		Processors: procs,
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

//todo make this function to method of Site
func Watch(s *Site) {
	cnf := s.SiteConfig
	processors := s.Processors
	filemods, err := Watcher(&cnf)
	errhandle(err)

	go func() {
		for {
			fn := <-filemods
			if !strings.HasPrefix(filepath.Base(fn), ".") {
				drainchannel(filemods)
				//TODO change it to site.Rerender()
				site := NewSite(&cnf, processors)
				site.Render()
			}
		}
	}()

}

func (site *Site) Collect() {
	errors := make(chan error, 10)

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

func (site *Site) Process() (int, error) {
	processed := 0
	for _, page := range site.Pages {
		if page.Changed() {
			debug("Processing page %s\n", page.Source)
			err := page.Process()
			if err != nil {
				return processed, err
			}
			processed++
		}
	}
	return processed, nil
}

func (site *Site) ProcessAll() error {
	for _, page := range site.Pages {
		err := page.Process()
		if err != nil {
			return err
		}
	}
	return nil
}

func (site *Site) Summary() {
	err := site.ProcessAll()
	errhandle(err)

	out("Total pages to render: %d\n", len(site.Pages))

	for _, page := range site.Pages {
		// do not output static files in summary mode
		if page.Rule == nil {
			continue
		}

		out("%s - %s: %d chars; %s\n",
			page.Path, page.Title, len(page.Content()), page.Rule)
		out("------------")
		_, err := page.WriteTo(os.Stdout)
		errhandle(err)
		out("------------\n")
	}
}

func (site *Site) Render() {
	processed, err := site.Process()
	errhandle(err)
	out("Rendering %d changed pages of %d total\n", processed, len(site.Pages))

	for _, page := range site.Pages {
		if !page.Changed() {
			continue
		}

		debug("Rendering %s -> %s\n", page.Source, page.OutputPath())

		err := os.MkdirAll(filepath.Dir(page.OutputPath()), 0755)
		errhandle(err)

		_, err = page.Render()
		if err != nil {
			errhandle(fmt.Errorf("Unable to render page '%s': %v", page.Source, err))
		}
	}
}

func (site *Site) Lookup(path string) *Page {
	for i := range site.Pages {
		if site.Pages[i].FullPath() == path {
			return site.Pages[i]
		}
	}
	return nil
}

func (site *Site) PageBySomePath(s string) *Page {
	if strings.HasPrefix(s, site.Source) {
		rel, err := filepath.Rel(site.Source, s)
		if err != nil {
			return nil
		}
		return site.Pages.BySource(rel)
	}
	if strings.HasPrefix(s, site.Output) {
		rel, err := filepath.Rel(site.Output, s)
		if err != nil {
			return nil
		}
		return site.Pages.ByPath(rel)
	}
	if page := site.Pages.BySource(s); page != nil {
		return page
	}
	return site.Pages.ByPath(s)
}

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
	ConfigPath string
	SiteConfig
	Template  *template.Template
	ChangedAt time.Time
	Pages     PageSlice

	ForceRefresh bool

	mx sync.Mutex

	Processors map[string]Processor
}

// make new instance of Site
func NewSite(configPath string, procs ProcessorMap) *Site {
	site := &Site{
		ConfigPath: configPath,
		//SiteConfig: *config,
		//Template:   template,
		//ChangedAt:  changed,
		//		Pages:      make(PageSlice, 0),
		Processors: procs,
	}

	site.Reconfig()

	return site
}

// read site config, templates and find all eligible pages
func (site *Site) Reconfig() {
	config, err := NewSiteConfig(site.ConfigPath)
	if err != nil {
		errhandle(fmt.Errorf("invalid config file '%s': %v", site.ConfigPath, err))
		os.Exit(2) // ExitCodeInvalidConfig
	}
	site.SiteConfig = *config

	template := template.New("no-idea-what-to-pass-here").Funcs(TemplateFuncMap)
	template, err = template.ParseFiles(site.SiteConfig.Templates...)
	errhandle(err)

	changed := site.SiteConfig.changedAt
	for _, fn := range site.SiteConfig.Templates {
		stat, err := os.Stat(fn)
		if err == nil && changed.Before(stat.ModTime()) {
			changed = stat.ModTime()
		}
	}

	site.Template = template
	site.ChangedAt = changed
	site.Pages = make(PageSlice, 0)

	site.Collect()
	site.FindDeps()
}

func (site *Site) AddPages(path string) {
	for _, page := range NewPages(site, path) {
		if page.state != StateIgnored {
			site.Pages = append(site.Pages, page)
		}
	}
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
			site.AddPages(fn)
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
			_, err := page.Process()
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
		_, err := page.Process()
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

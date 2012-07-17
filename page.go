// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// stuff to remember
// - научиться парсить дату из файла (?) - забить пока не заработает всë
// - решить что-то с тегами

type Page struct {
	PageConfig

	Site		 *Site
	Pattern		 string
	Rules		 []string

	Content		 string
	Path		 string
	ModTime		 time.Time
	RenderTime	 time.Time
}

type PageSlice []*Page

func NewPage(site *Site, path string) *Page {
	text, err := ioutil.ReadFile(path)
	errhandle(err)

	stat, err := os.Stat(path)
	errhandle(err)

	relpath, err := filepath.Rel(site.Path, path)
	errhandle(err)

	pattern, rules := site.Rules.MatchedRules(path)

	page := &Page{
		Site:    site,
		Pattern: pattern,
		Rules:   rules,
		Content: string(text),
		Path:    relpath,
		ModTime: stat.ModTime(),
	}
	page.ReadConfig()
	return page
}

func (page *Page) ReadConfig() {
	if page.Rules == nil || page.Rules[0] == ":ignore" {
		return
	}

	parts := strings.SplitN(page.Content, "----\n", 2)
	if len(parts) == 2 {
		page.PageConfig = *ParseConfig(parts[0])
		page.Content = parts[1]
	} else {
		page.PageConfig = *ParseConfig("")
	}
}

func (page *Page) Process() {
	for _, rule := range page.Rules {
		ProcessRule(page, rule)
	}
}

func (page *Page) Url() string {
	return strings.Replace(page.Path, "/index.html", "/", 1)
}

// PageSlice manipulation

func (pages PageSlice) Get(i int) *Page { return pages[i] }
func (pages PageSlice) First() *Page    { return pages.Get(0) }
func (pages PageSlice) Last() *Page     { return pages.Get(len(pages) - 1) }

func (pages PageSlice) Len() int {
	return len(pages)
}
func (pages PageSlice) Less(i, j int) bool {
	return pages.Get(i).ModTime.Unix() < pages.Get(j).ModTime.Unix()
}
func (pages PageSlice) Swap(i, j int) {
	pages[i], pages[j] = pages[j], pages[i]
}

func (pages PageSlice) Sort() {
	sort.Sort(pages)
}

func (pages PageSlice) Children(root string) *PageSlice {
	children := make(PageSlice, 0)

	for _, page := range pages {
		if strings.HasPrefix(page.Url(), root) && page.Url() != root {
			children = append(children, page)
		}
	}

	return &children
}

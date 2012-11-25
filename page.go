// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Page struct {
	PageHeader

	Site      *Site
	Rule      *Rule
	Pattern   string
	Processed bool

	Content    string
	Source     string
	Path       string
	ModTime    time.Time
}

type PageSlice []*Page

func NewPage(site *Site, path string) *Page {
	stat, err := os.Stat(path)
	errhandle(err)

	relpath, err := filepath.Rel(site.Source, path)
	errhandle(err)

	pattern, rule := site.Rules.MatchedRule(relpath)

	page := &Page{
		Site:      site,
		Rule:      rule,
		Pattern:   pattern,
		Processed: false,
		Content:   "",
		Source:    relpath,
		Path:      relpath,
		ModTime:   stat.ModTime(),
	}
	page.Peek()
	return page
}

func (page *Page) GetContent() string {
	if len(page.Content) == 0 {
		content, err := ioutil.ReadFile(page.FullPath())
		errhandle(err)
		page.SetContent(string(content))
	}
	return page.Content
}

func (page *Page) SetContent(content string) {
	page.Content = content
}

func (page *Page) FullPath() string {
	return filepath.Join(page.Site.Source, page.Source)
}

func (page *Page) Url() string {
	return strings.Replace(page.Path, "/index.html", "/", 1)
}

func (page *Page) UrlTo(other *Page) string {
       return page.Rel(other.Url())
}

func (page *Page) Rel(path string) string {
       root := strings.Repeat("../", strings.Count(page.Url(), "/"))
       return root + path
}

// Peek is used to run those processors which should be done before others can
// find out about us. Two actual examples include 'config' and 'rename'
// processors right now.
func (page *Page) Peek() {
	for _, name := range PreProcessors {
		cmd := page.Rule.MatchedCommand(name)
		if cmd != nil {
			ProcessCommand(page, cmd)
		}
	}
}

func (page *Page) Process() {
	if page.Processed {
		return
	}

	page.Processed = true
	if page.Rule.Commands != nil {
		for _, cmd := range page.Rule.Commands {
			if !cmd.MatchesAny(PreProcessors) {
				ProcessCommand(page, &cmd)
			}
		}
	}
}

func (page *Page) WriteTo(writer io.Writer) (n int64, err error) {
	if !page.Processed {
		page.Process()
	}

	if page.Rule.Commands == nil {
		file, err := os.Open(page.FullPath())
		if err != nil {
			n = 0
		} else {
			n, err = io.Copy(writer, file)
		}
	} else {
		nint, werr := writer.Write([]byte(page.Content))
		n = int64(nint)
		err = werr
	}
	return n, err
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

func (pages PageSlice) HasPage(check func (page *Page) bool) bool {
	for _, page := range pages {
		if check(page) {
			return true
		}
	}
	return false
}

func (pages PageSlice) WithTag(tag string) *PageSlice {
	tagged := make(PageSlice, 0)

	for _, page := range pages {
		if page.Tags != nil && SliceStringIndexOf(page.Tags, tag) != -1 {
			tagged = append(tagged, page)
		}
	}

	return &tagged
}

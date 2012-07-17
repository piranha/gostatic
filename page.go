package main

import (
	"os"
	"time"
	"bytes"
	"io/ioutil"
	"text/template"
	"path/filepath"
	"strings"
	blackfriday "github.com/russross/blackfriday"
	"sort"
)

type Page struct {
	Site *Site
	Config
	Content    string
	Path       string
	ModTime    time.Time
	RenderTime time.Time
}

type PageSlice []*Page

func NewPage(site *Site, path string) *Page {
	text, err := ioutil.ReadFile(path)
	errhandle(err)

	stat, err := os.Stat(path)
	errhandle(err)

	relpath, err := filepath.Rel(site.Path, path)
	errhandle(err)

	head, content := SplitHead(text)

	return &Page{
		Site:    site,
		Config:  *ParseConfig(head),
		Content: content,
		Path:    relpath,
		ModTime: stat.ModTime(),
	}
}

// Destination() returns relative path to a file in future published repository
func (page *Page) Destination() string {
	path := page.Path

	if strings.HasSuffix(path, ".html") {
		return path
	}

	if strings.HasSuffix(path, ".md") {
		path = strings.Replace(path, ".md", ".html", 1)
	}

	if filepath.Base(path) != "index.html" {
		path = strings.Replace(path, ".html", "/index.html", 1)
	}

	return path
}

func (page *Page) Url() string {
	return strings.Replace(page.Destination(), "/index.html", "/", 1)
}

func (page *Page) Rendered() []byte {
	ctmpl, err := template.New("ad-hoc").Parse(page.Content)
	errhandle(err)

	var buffer bytes.Buffer
	err = ctmpl.Execute(&buffer, page)
	errhandle(err)

	return blackfriday.MarkdownCommon(buffer.Bytes())
}

func SplitHead(text []byte) (string, string) {
	parts := bytes.SplitN(text, []byte("----\n"), 2)
	return string(parts[0]), string(parts[1])
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

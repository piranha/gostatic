package main

import (
	"os"
	"time"
	"bytes"
	"io/ioutil"
	"text/template"
	"path/filepath"
	blackfriday "github.com/russross/blackfriday"
)

type Page struct {
	Site *Site
	Config
	Content string
	Path string
	Mod time.Time
}

func NewPage(site *Site, path string) *Page {
	text, err := ioutil.ReadFile(path)
	errhandle(err)

	stat, err := os.Stat(path)
	errhandle(err)

	relpath, err := filepath.Rel(site.Path, path)
	errhandle(err)

	head, content := SplitHead(text)

	return &Page{
		Site: site,
		Config: *ParseConfig(head),
		Content: content,
		Path: relpath,
		Mod: stat.ModTime(),
	}
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

// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"os"
	"time"
	"fmt"
	goopt "github.com/droundy/goopt"
	blackfriday "github.com/russross/blackfriday"
	"text/template"
	"io/ioutil"
	"bytes"
	"path/filepath"
)

var Version = "0.1"

var Summary = `gostatic -t template [-t template] sitedir

Build a site
`

var templates = goopt.Strings([]string{"-t", "--template"},
	"template", "path to template")
var showVersion = goopt.Flag([]string{"-v", "--version"}, []string{},
	"show version and exit", "")


func main() {
	goopt.Version = Version
	goopt.Summary = Summary

	goopt.Parse(nil)

	if *showVersion {
		fmt.Printf("gostatic %s\n", goopt.Version)
		return
	}

	if len(*templates) == 0 || len(goopt.Args) == 0 {
		println(goopt.Usage())
		return
	}

	t, err := template.ParseFiles(*templates...)
	errhandle(err)

	site := NewSite(t, goopt.Args[0])
	site.Summary()
}


type Page struct {
	Config
	Content []byte
	Path string
	Mod time.Time
	Template *template.Template
}

func NewPage(path string, base string, t *template.Template) *Page {
	text, err := ioutil.ReadFile(path)
	errhandle(err)

	stat, err := os.Stat(path)
	errhandle(err)

	head, content := SplitHead(text)

	return &Page{
		Config: *ParseConfig(string(head)),
		Content: blackfriday.MarkdownCommon(content),
		Path: path,
		Mod: stat.ModTime(),
		Template: t,
	}
}

func SplitHead(text []byte) ([]byte, []byte) {
	parts := bytes.SplitN(text, []byte("----\n"), 2)
	return parts[0], parts[1]
}


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

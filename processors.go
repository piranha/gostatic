package main

import (
	"bytes"
	"errors"
	blackfriday "github.com/russross/blackfriday"
	"path/filepath"
	"strings"
	"text/template"
)

type Processor func(page *Page, args []string)

var Processors = map[string]Processor{
	":inner-template": ProcessInnerTemplate,
	":template":       ProcessTemplate,
	":markdown":       ProcessMarkdown,
	":rename":         ProcessRename,
	":ignore":         ProcessIgnore,
	":directorify":    ProcessDirectorify,
}

func ProcessRule(page *Page, rule string) {
	bits := strings.Split(rule, " ")
	if strings.HasPrefix(bits[0], ":") { // internal processing
		processor := Processors[bits[0]]
		processor(page, bits[1:])
	} else { // external processing
		println("no idea", rule)
	}
}

func ProcessInnerTemplate(page *Page, args []string) {
	t, err := template.New("ad-hoc").Parse(page.Content)
	errhandle(err)

	var buffer bytes.Buffer
	err = t.Execute(&buffer, page)
	errhandle(err)

	page.Content = buffer.String()
}

func ProcessTemplate(page *Page, args []string) {
	var pagetype string
	if len(args) > 0 {
		pagetype = args[0]
	} else {
		pagetype = page.Type
	}

	var buffer bytes.Buffer
	err := page.Site.Template.ExecuteTemplate(&buffer, pagetype, page)
	errhandle(err)

	page.Content = buffer.String()
}

func ProcessMarkdown(page *Page, args []string) {
	result := blackfriday.MarkdownCommon([]byte(page.Content))
	page.Content = string(result)
}

func ProcessRename(page *Page, args []string) {
	if len(args) < 1 {
		errhandle(errors.New(":rename rule needs an argument"))
	}
	dest := strings.Replace(args[0], "*", "", -1)
	pattern := strings.Replace(page.Pattern, "*", "", -1)

	page.Path = strings.Replace(page.Path, pattern, dest, -1)
}

func ProcessIgnore(page *Page, args []string) {
	page.Path = ""
}

func ProcessDirectorify(page *Page, args []string) {
	if filepath.Base(page.Path) != "index.html" {
		page.Path = strings.Replace(page.Path, ".html", "/index.html", 1)
	}
}

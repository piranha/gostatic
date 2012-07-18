package main

import (
	"bytes"
	"errors"
	blackfriday "github.com/russross/blackfriday"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Processor func(page *Page, args []string)

var Processors = map[string]Processor{
	"inner-template": ProcessInnerTemplate,
	"template":       ProcessTemplate,
	"markdown":       ProcessMarkdown,
	"rename":         ProcessRename,
	"ignore":         ProcessIgnore,
	"directorify":    ProcessDirectorify,
	"external":       ProcessExternal,
}

func ProcessRule(page *Page, rule string) {
	if strings.HasPrefix(rule, ":") {
		rule = "external " + rule[1:]
	}
	bits := strings.Split(rule, " ")
	processor := Processors[bits[0]]
	processor(page, bits[1:])
}

func ProcessInnerTemplate(page *Page, args []string) {
	t, err := template.New("ad-hoc").Parse(page.GetContent())
	errhandle(err)

	var buffer bytes.Buffer
	err = t.Execute(&buffer, page)
	errhandle(err)

	page.SetContent(buffer.String())
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

	page.SetContent(buffer.String())
}

func ProcessMarkdown(page *Page, args []string) {
	result := blackfriday.MarkdownCommon([]byte(page.GetContent()))
	page.SetContent(string(result))
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

func ProcessExternal(page *Page, args []string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(page.GetContent())
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	errhandle(err)

	page.SetContent(out.String())
}

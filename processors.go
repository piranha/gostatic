package main

import (
	"bytes"
	"errors"
	blackfriday "github.com/russross/blackfriday"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"fmt"
)

type Processor struct {
	Func func(page *Page, args []string)
	Desc string
}

var Processors = map[string]*Processor{
	"inner-template": &Processor{
		ProcessInnerTemplate,
		"process content as a Go template",
	},
	"template":       &Processor{
		ProcessTemplate,
		"put content in a template (by default in 'page' template)",
	},
	"markdown":       &Processor{
		ProcessMarkdown,
		"process content as a markdown",
	},
	"rename":         &Processor{
		ProcessRename,
		"rename resulting file",
	},
	"ignore":         &Processor{
		ProcessIgnore,
		"ignore file",
	},
	"directorify":    &Processor{
		ProcessDirectorify,
		"path/name.html -> path/name/index.html",
	},
	"external":       &Processor{
		ProcessExternal,
		"run external command to process content (shortcut ':')",
	},
}

func ProcessorSummary() {
	for k, p := range Processors {
		fmt.Printf("%s\n\t%s\n", k, p.Desc)
	}
}

func ProcessRule(page *Page, rule string) {
	if strings.HasPrefix(rule, ":") {
		rule = "external " + rule[1:]
	}
	bits := strings.Split(rule, " ")
	processor := Processors[bits[0]]
	processor.Func(page, bits[1:])
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

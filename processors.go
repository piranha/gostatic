package main

import (
	"bytes"
	"errors"
	"fmt"
	blackfriday "github.com/russross/blackfriday"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type Processor struct {
	Func func(page *Page, args []string)
	Desc string
}

// PreProcessors is a list of processor necessary to be executed beforehand to
// fill out information, which can be required by fellow pages
var PreProcessors = RuleList{"config", "rename", "directorify"}

var Processors = map[string]*Processor{
	"inner-template": &Processor{
		ProcessInnerTemplate,
		"process content as a Go template",
	},
	"template": &Processor{
		ProcessTemplate,
		"put content in a template (by default in 'page' template)",
	},
	"markdown": &Processor{
		ProcessMarkdown,
		"process content as a markdown",
	},
	"rename": &Processor{
		ProcessRename,
		"rename resulting file",
	},
	"ignore": &Processor{
		ProcessIgnore,
		"ignore file",
	},
	"directorify": &Processor{
		ProcessDirectorify,
		"path/name.html -> path/name/index.html",
	},
	"external": &Processor{
		ProcessExternal,
		"run external command to process content (shortcut ':')",
	},
	"config": &Processor{
		ProcessConfig,
		"read config from content (separated by '----\\n')",
	},
}

func ProcessorSummary() {
	keys := make([]string, 0)
	for k := range Processors {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		p := Processors[k]
		pre := ""
		if PreProcessors.MatchedIndex(k) != -1 {
			pre = "(preprocessor)"
		}
		fmt.Printf("%s %s\n\t%s\n", k, pre, p.Desc)
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
	var idx int
	site := page.Site
	for i, pg := range site.Pages {
		if pg == page {
			idx = i
			break
		}
	}
	site.Pages = append(site.Pages[:idx], site.Pages[idx + 1:]...)
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

func ProcessConfig(page *Page, args []string) {
	parts := strings.SplitN(page.GetContent(), "----\n", 2)
	if len(parts) != 2 {
		errhandle(errors.New(fmt.Sprintf(
			"page %s has no configuration in the head, while it is " +
			"requested by the site configuration",
			page.Path)))
	}

	page.PageConfig = *ParseConfig(string(parts[0]))
	page.SetContent(parts[1])
}

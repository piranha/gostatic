package main

import (
	"bytes"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

type Processor struct {
	Func func(page *Page, args []string)
	Desc string
}

// PreProcessors is a list of processor necessary to be executed beforehand to
// fill out information, which can be required by fellow pages
var PreProcessors = CommandList{"config", "rename", "ext", "directorify",
	"tags", "ignore"}

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
		"rename resulting file (argument - pattern for renaming)",
	},
	"ext": &Processor{
		ProcessExt,
		"change extension",
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
	"tags": &Processor{
		ProcessTags,
		("generate tags pages for tags mentioned in page header " +
			"(argument - tag template)"),
	},
	"relativize": &Processor{
		ProcessRelativize,
		"Relativize URLs.",
	},
}

func ProcessRelativize(page *Page, args []string) {
	//reg, err := regexp.Compile(`\[(.*)\]\(/(.*)\)`)
	reg, err := regexp.Compile(`<a href=\"/(.*)\">(.*)</a>`)
	if err != nil {
		errhandle(fmt.Errorf("%s: %s", page.Source, err))
	}
	//repl := "[$1](" + page.Rel("/") + "$2)"
	repl := `<a href="` + page.Rel("/") + `$1">$2</a>`
	relativized := reg.ReplaceAllString(page.Content(), repl)
	page.SetContent(relativized)
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
		if PreProcessors.MatchedIndex(Command(k)) != -1 {
			pre = "(preprocessor)"
		}
		fmt.Printf("%s %s\n\t%s\n", k, pre, p.Desc)
	}
}

func ProcessCommand(page *Page, cmd *Command) {
	c := string(*cmd)
	if strings.HasPrefix(c, ":") {
		c = "external " + c[1:]
	}
	bits := strings.Split(c, " ")
	processor := Processors[bits[0]]
	if processor == nil {
		errhandle(fmt.Errorf("processor '%s' not found", bits[0]))
	}
	processor.Func(page, bits[1:])
}

func ProcessInnerTemplate(page *Page, args []string) {
	t, err := page.Site.Template.Clone()
	errhandle(err)
	t, err = t.New("ad-hoc").Parse(page.Content())
	if err != nil {
		errhandle(fmt.Errorf("Page %s: %s", page.Source, err))
	}

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, "ad-hoc", page)
	if err != nil {
		errhandle(fmt.Errorf("Page %s: %s", page.Source, err))
	}

	page.SetContent(buffer.String())
}

func ProcessTemplate(page *Page, args []string) {
	if len(args) < 1 {
		errhandle(errors.New("'template' rule needs an argument"))
	}
	pagetype := args[0]

	var buffer bytes.Buffer
	err := page.Site.Template.ExecuteTemplate(&buffer, pagetype, page)
	if err != nil {
		errhandle(fmt.Errorf("%s: %s", page.Source, err))
	}

	page.SetContent(buffer.String())
}

func ProcessMarkdown(page *Page, args []string) {
	result := Markdown(page.Content())
	page.SetContent(result)
}

func ProcessRename(page *Page, args []string) {
	if len(args) < 1 {
		errhandle(errors.New("'rename' rule needs an argument"))
	}
	dest := strings.Replace(args[0], "*", "", -1)
	pattern := strings.Replace(page.Pattern, "*", "", -1)

	page.Path = strings.Replace(page.Path, pattern, dest, -1)
}

func ProcessExt(page *Page, args []string) {
	if len(args) < 1 {
		errhandle(errors.New(
			"'ext' rule requires an extension prefixed with dot"))
	}
	ext := filepath.Ext(page.Path)
	if ext == "" {
		page.Path = page.Path + args[0]
	} else {
		page.Path = strings.Replace(page.Path, ext, args[0], 1)
	}
}

func ProcessIgnore(page *Page, args []string) {
	page.state = StateIgnored
}

func ProcessDirectorify(page *Page, args []string) {
	if filepath.Base(page.Path) != "index.html" {
		page.Path = strings.Replace(page.Path, ".html", "/index.html", 1)
	}
}

func ProcessExternal(page *Page, args []string) {
	path, err := exec.LookPath(args[0])
	if err != nil {
		path, err = exec.LookPath(filepath.Join(page.Site.Base, args[0]))
		if err != nil {
			errhandle(fmt.Errorf(
				"command '%s' not found", args[0]))
		}
	}
	cmd := exec.Command(path, args[1:]...)
	cmd.Stdin = strings.NewReader(page.Content())
	cmd.Dir = page.Site.Base
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		errhandle(fmt.Errorf("Error executing '%s': %s\n%s",
			strings.Join(args, " "), err, stderr.String()))
	}

	page.SetContent(string(data))
}

func ProcessConfig(page *Page, args []string) {
	parts := strings.SplitN(page.Content(), "----\n", 2)
	if len(parts) != 2 {
		// no configuration, well then...
		page.PageHeader = *NewPageHeader()
		return
	}

	page.PageHeader = *ParseHeader(parts[0])
	page.SetContent(parts[1])
}

func ProcessTags(page *Page, args []string) {
	if len(args) < 1 {
		errhandle(errors.New("'tags' rule needs an argument"))
	}

	if page.Tags == nil {
		return
	}

	site := page.Site

	for _, tag := range page.Tags {
		if !site.Pages.HasPage(func(inner *Page) bool {
			return inner.Title == tag
		}) {
			path := strings.Replace(args[0], "*", tag, 1)
			pattern, rule := site.Rules.MatchedRule(path)
			tagpage := &Page{
				PageHeader: PageHeader{Title: tag},
				Site:       site,
				Pattern:    pattern,
				Rule:       rule,
				Source:     path,
				Path:       path,
				// tags are never new, because they only depend on pages and
				// have not a bit of original content
				ModTime: time.Unix(0, 0),
			}
			page.Site.Pages = append(page.Site.Pages, tagpage)
			tagpage.peek()
		}
	}

}

// template utilities

var inventory = map[string]interface{}{}

func HasChanged(name string, value interface{}) bool {
	changed := true

	if inventory[name] == value {
		changed = false
	} else {
		inventory[name] = value
	}

	return changed
}

func Cut(value, begin, end string) (string, error) {
	bre, err := regexp.Compile(begin)
	if err != nil {
		return "", err
	}
	ere, err := regexp.Compile(end)
	if err != nil {
		return "", err
	}

	bloc := bre.FindIndex([]byte(value))
	eloc := ere.FindIndex([]byte(value))

	return value[bloc[1]:eloc[0]], nil
}

func Hash(value string) string {
	h := adler32.New()
	io.WriteString(h, value)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Versionize(current *Page, value string) string {
	page := current.Site.Pages.ByPath(value)
	c := page.Process().Content()
	h := Hash(c)
	return current.UrlTo(page) + "?v=" + h
}

var funcMap = template.FuncMap{
	"changed": HasChanged,
	"cut":     Cut,
	"hash":    Hash,
	"version": Versionize,
}

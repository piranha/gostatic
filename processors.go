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

var Processors map[string]*Processor

// it is necessary to wrap assignment in a function since go compiler is strict
// enough to throw error when there is an assignment loop, but not smart enough
// to determine if it's actually truth (Processors definition loops with
// ProcessTags, which is strange)
func InitProcessors() {
	Processors = map[string]*Processor{
		"inner-template": &Processor{
			ProcessInnerTemplate,
			"process content as a Go template",
		},
		"template": &Processor{
			ProcessTemplate,
			"put content in a template (argument - template name)",
		},
		"markdown": &Processor{
			ProcessMarkdown,
			"process content as a markdown",
		},
		"rename": &Processor{
			ProcessRename,
			"rename resulting file (argument - pattern for renaming, " +
				"relative to current file location)",
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
			("make all urls bound at root relative " +
				"(allows deploying resulting site in a subdirectory)"),
		},
	}
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
		errhandle(fmt.Errorf("'rename' rule needs an argument"))
	}

	dest := args[0]

	if strings.Contains(dest, "*") {
		if !strings.Contains(page.Pattern, "*") {
			errhandle(fmt.Errorf(
				"'rename' rule cannot rename '%s' to '%s'",
				page.Pattern, dest))
		}

		group := fmt.Sprintf("([^%c]*)", filepath.Separator)
		base := filepath.Base(page.Pattern)
		pat := strings.Replace(regexp.QuoteMeta(base), "\\*", group, 1)

		re, err := regexp.Compile(pat)
		errhandle(err)
		m := re.FindStringSubmatch(filepath.Base(page.Path))

		dest = strings.Replace(dest, "*", m[1], 1)
	}

	page.Path = filepath.Join(filepath.Dir(page.Path), dest)
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
		errhandle(fmt.Errorf("'%s' failed: %s\n%s",
			strings.Join(args, " "), err, stderr.String()))
	}

	page.SetContent(string(data))
}

func ProcessConfig(page *Page, args []string) {
	parts := TrimSplitN(page.Content(), "\n----\n", 2)
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
		path := strings.Replace(args[0], "*", tag, 1)

		if !site.Pages.HasPage(func(inner *Page) bool {
			return inner.Source == path
		}) {
			pattern, rule := site.Rules.MatchedRule(path)
			tagpage := &Page{
				PageHeader: PageHeader{Title: tag},
				Site:       site,
				Pattern:    pattern,
				Rule:       rule,
				Source:     path,
				Path:       path,
				wasread:    true,
				// tags are never new, because they only depend on pages and
				// have not a bit of original content
				ModTime: time.Unix(0, 0),
			}
			page.Site.Pages = append(page.Site.Pages, tagpage)
			tagpage.peek()
		}
	}

}

var RelRe = regexp.MustCompile(`(href|src)=["']/([^"']*)["']`)
var NonProtoRe = regexp.MustCompile(`(href|src)=["']//`)

func ProcessRelativize(page *Page, args []string) {
	repl := `$1="` + page.Rel("/") + `$2"`
	rv := RelRe.ReplaceAllStringFunc(page.Content(), func(inp string) string {
		if NonProtoRe.MatchString(inp) {
			return inp
		}
		return RelRe.ReplaceAllString(inp, repl)
	})
	page.SetContent(rv)
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

	if bloc == nil {
		bloc = []int{0, 0}
	}
	if eloc == nil {
		eloc = []int{len(value)}
	}

	return value[bloc[1]:eloc[0]], nil
}

func Hash(value string) string {
	h := adler32.New()
	io.WriteString(h, value)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Versionize(current *Page, value string) string {
	page := current.Site.Pages.ByPath(value)
	if page == nil {
		errhandle(fmt.Errorf(
			"trying to versionize page which does not exist: %s, current: %s",
			value, current.Path))
	}
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

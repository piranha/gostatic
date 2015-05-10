package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"strconv"
	"time"
)

type Processor struct {
	Func func(page *Page, args []string)
	Desc string
	// Preprocessor is a processor which will be executed during initialization
	// stage
	Pre bool
}

var Processors map[string]*Processor

// it is necessary to wrap assignment in a function since go compiler is strict
// enough to throw error when there can be an assignment loop, but not smart
// enough to determine if it's actually truth (Processors definition loops with
// ProcessTags, which is strange)
func InitProcessors() {
	Processors = map[string]*Processor{
		"inner-template": &Processor{
			ProcessInnerTemplate,
			"process content as a Go template",
			false,
		},
		"template": &Processor{
			ProcessTemplate,
			"put content in a template (argument - template name)",
			false,
		},
		"markdown": &Processor{
			ProcessMarkdown,
			"process content as a markdown",
			false,
		},
		"rename": &Processor{
			ProcessRename,
			"rename resulting file (argument - pattern for renaming, " +
				"relative to current file location)",
			true,
		},
		"ext": &Processor{
			ProcessExt,
			"change extension",
			true,
		},
		"ignore": &Processor{
			ProcessIgnore,
			"ignore file",
			true,
		},
		"directorify": &Processor{
			ProcessDirectorify,
			"path/name.html -> path/name/index.html",
			false,
		},
		"external": &Processor{
			ProcessExternal,
			"run external command to process content (shortcut ':')",
			false,
		},
		"config": &Processor{
			ProcessConfig,
			"read config from content (separated by '----\\n')",
			true,
		},
		"tags": &Processor{
			ProcessTags,
			"generate tags pages for tags mentioned in page header " +
				"(argument - tag template)",
			true,
		},
		"relativize": &Processor{
			ProcessRelativize,
			"make all urls bound at root relative " +
				"(allows deploying resulting site in a subdirectory)",
			false,
		},
		"paginate": &Processor{
			ProcessPaginate,
			"partition lists of pages " +
				"(arguments - amount to partition by, list page template)",
			true,
		},
	}
}

func ProcessorSummary() {
	keys := make([]string, 0, len(Processors))
	for k := range Processors {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		p := Processors[k]
		pre := ""
		if p.Pre {
			pre = "(preprocessor)"
		}
		fmt.Printf("%s %s\n\t%s\n", k, pre, p.Desc)
	}
}

func ProcessCommand(page *Page, cmd *Command, pre bool) {
	c := string(*cmd)
	if strings.HasPrefix(c, ":") {
		c = "external " + c[1:]
	}
	bits := strings.Split(c, " ")

	processor := Processors[bits[0]]
	if processor.Pre != pre {
		return
	}
	if processor == nil {
		errhandle(fmt.Errorf("processor '%s' not found", bits[0]))
	}
	processor.Func(page, bits[1:])
}

func ProcessInnerTemplate(page *Page, args []string) {
	defer func() {
		if err := recover(); err != nil {
			errhandle(fmt.Errorf("%s: %s", page.Source, err))
		}
	}()

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
		errexit(errors.New("'template' rule needs an argument"))
	}
	pagetype := args[0]

	defer func() {
		if err := recover(); err != nil {
			errhandle(fmt.Errorf("%s: %s", page.Source, err))
		}
	}()

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
		errexit(fmt.Errorf("'rename' rule needs an argument"))
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
		errexit(errors.New(
			"'ext' rule requires an extension prefixed with dot"))
	}
	newExt := args[0]

	ext := filepath.Ext(page.Path)
	if ext == "" {
		page.Path = page.Path + newExt
	} else {
		page.Path = page.Path[0:len(page.Path) - len(ext)] + newExt
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
	if len(args) < 1 {
		errexit(errors.New("'external' rule needs a command name"))
	}
	cmdName := args[0]
	cmdArgs := args[1:]

	path, err := exec.LookPath(cmdName)
	if err != nil {
		path, err = exec.LookPath(filepath.Join(page.Site.Base, cmdName))
		if err != nil {
			errhandle(fmt.Errorf("command '%s' not found", cmdName))
		}
	}

	cmd := exec.Command(path, cmdArgs...)
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
		errexit(errors.New("'tags' rule needs an argument"))
	}
	pathTemplate := args[0]

	if page.Tags == nil {
		return
	}

	site := page.Site

	for _, tag := range page.Tags {
		tagpath := strings.Replace(pathTemplate, "*", tag, 1)

		if site.Pages.BySource(tagpath) == nil {
			pattern, rule := site.Rules.MatchedRule(tagpath)
			if rule == nil {
				errexit(fmt.Errorf("Tag path '%s' does not match any rule",
					tagpath))
			}
			tagpage := &Page{
				PageHeader: PageHeader{Title: tag},
				Site:       site,
				Pattern:    pattern,
				Rule:       rule,
				Source:     tagpath,
				Path:       tagpath,
				wasread:    true,
				// tags are never new, because they only depend on pages and
				// have not a bit of original content
				ModTime:    time.Unix(0, 0),
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

var PageCounter = map[string]int{}
var Paginated = map[string]PageSlice{}

func ProcessPaginate(page *Page, args []string) {
	if len(args) < 2 {
		errexit(errors.New("'paginate' rule needs two arguments"))
	}
	length, err := strconv.Atoi(args[0])
	if err != nil { errexit(err) }
	pathTemplate := args[1]

	if val, ok := PageCounter[pathTemplate]; ok {
		PageCounter[pathTemplate] = val + 1
	} else {
		PageCounter[pathTemplate] = 1
	}

	site := page.Site

	// page number, 1-based
	n := strconv.Itoa(1 + ((PageCounter[pathTemplate] - 1) / length))
	println(page.Source, n)
	listpath := strings.Replace(pathTemplate, "*", n, 1)
	listpage := site.Pages.BySource(listpath)

	if listpage == nil {
		pattern, rule := site.Rules.MatchedRule(listpath)
		if rule == nil {
			errexit(fmt.Errorf("Paginated path '%s' does not match any rule",
				listpath))
		}
		listpage = &Page{
			PageHeader: PageHeader{Title: n},
			Site:		site,
			Pattern:	pattern,
			Rule:		rule,
			Source:		listpath,
			Path:		listpath,
			wasread:	true,
			ModTime:	time.Unix(0, 0),
		}
		page.Site.Pages = append(page.Site.Pages, listpage)
		listpage.peek()

		Paginated[listpath] = make(PageSlice, 0)
	}

	Paginated[listpath] = append(Paginated[listpath], page)
}

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// Preprocessor is a processor which will be executed during initialization
	// stage
	Pre    = 1 << iota
	Hidden = 1 << iota
)

type Processor struct {
	Func func(page *Page, args []string)
	Desc string
	Mode int
}

var Processors map[string]*Processor

// it is necessary to wrap assignment in a function since go compiler is strict
// enough to throw error when there may be an assignment loop, but not smart
// enough to determine if there is one (Processors definition loops with
// ProcessTags)
func InitProcessors() {
	Processors = map[string]*Processor{
		"inner-template": &Processor{
			ProcessInnerTemplate,
			"process content as a Go template",
			0,
		},
		"template": &Processor{
			ProcessTemplate,
			"put content in a template (argument - template name)",
			0,
		},
		"markdown": &Processor{
			ProcessMarkdown,
			"process content as a markdown",
			0,
		},
		"rename": &Processor{
			ProcessRename,
			"rename resulting file (argument - pattern for renaming, " +
				"relative to current file location)",
			Pre,
		},
		"ext": &Processor{
			ProcessExt,
			"change extension",
			Pre,
		},
		"ignore": &Processor{
			ProcessIgnore,
			"ignore file",
			Pre,
		},
		"directorify": &Processor{
			ProcessDirectorify,
			"path/name.html -> path/name/index.html",
			Pre,
		},
		"external": &Processor{
			ProcessExternal,
			"run external command to process content (shortcut ':')",
			0,
		},
		"config": &Processor{
			ProcessConfig,
			"read config from content (separated by '----\\n')",
			Pre,
		},
		"tags": &Processor{
			ProcessTags,
			"generate tags pages for tags mentioned in page header " +
				"(argument - tag template)",
			Pre,
		},
		"relativize": &Processor{
			ProcessRelativize,
			"make all urls bound at root relative " +
				"(allows deploying resulting site in a subdirectory)",
			0,
		},
		"paginate": &Processor{
			ProcessPaginate,
			"partition lists of pages " +
				"(arguments - amount to partition by, list page template)",
			Pre,
		},
		"paginate-collect-pages": &Processor{
			ProcessPaginateCollectPages,
			"collects pages for paginator",
			Hidden,
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
		if p.Mode&Hidden != 0 {
			continue
		}
		pre := ""
		if p.Mode&Pre != 0 {
			pre = "(preprocessor)"
		}
		fmt.Printf("- %s %s\n\t%s\n", k, pre, p.Desc)
	}
}

func ProcessCommand(page *Page, cmd *Command, pre bool) {
	c := string(*cmd)
	if strings.HasPrefix(c, ":") {
		c = "external " + c[1:]
	}
	bits := strings.Split(c, " ")

	processor := Processors[bits[0]]
	if (processor.Mode&Pre != 0) != pre {
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
		page.Path = page.Path[0:len(page.Path)-len(ext)] + newExt
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
	pathPattern := args[0]

	if page.Tags == nil {
		return
	}

	site := page.Site

	for _, tag := range page.Tags {
		tagpath := strings.Replace(pathPattern, "*", tag, 1)

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

//================ Pagination start

type Paginator struct {
	Number      int
	PathPattern string
	Page        *Page
	Pages       PageSlice
}

var Paginated = map[string]PageSlice{}
var Paginators = map[string]*Paginator{}

func (pagi Paginator) Prev() *Paginator {
	src := strings.Replace(pagi.PathPattern, "*", strconv.Itoa(pagi.Number-1), 1)
	if prev, ok := Paginators[src]; ok {
		return prev
	}
	return nil
}

func (pagi Paginator) Next() *Paginator {
	src := strings.Replace(pagi.PathPattern, "*", strconv.Itoa(pagi.Number+1), 1)
	if next, ok := Paginators[src]; ok {
		return next
	}
	return nil
}

func ProcessPaginate(page *Page, args []string) {
	if len(args) < 2 {
		errexit(errors.New("'paginate' rule needs two arguments"))
	}
	length, err := strconv.Atoi(args[0])
	if err != nil {
		errexit(err)
	}
	pathPattern := args[1]

	if pages, ok := Paginated[pathPattern]; ok {
		Paginated[pathPattern] = append(pages, page)
	} else {
		Paginated[pathPattern] = PageSlice{page}
	}

	site := page.Site

	// page number, 1-based
	n := 1 + ((len(Paginated[pathPattern]) - 1) / length)
	listpath := strings.Replace(pathPattern, "*", strconv.Itoa(n), 1)
	listpage := site.Pages.BySource(listpath)

	if listpage != nil {
		return
	}

	pattern, rule := site.Rules.MatchedRule(listpath)
	if rule == nil {
		errexit(fmt.Errorf("Paginators path '%s' does not match any rule",
			listpath))
	}

	if !strings.HasPrefix(string(rule.Commands[0]), "paginate-collect-pages") {
		rule.Commands = append(
			CommandList{Command("paginate-collect-pages " + args[0])},
			rule.Commands...)
	}

	listpage = &Page{
		PageHeader: PageHeader{Title: strconv.Itoa(n)},
		Site:       site,
		Pattern:    pattern,
		Rule:       rule,
		Source:     listpath,
		Path:       listpath,
		wasread:    true,
		ModTime:    time.Unix(int64(n), 0),
	}
	page.Site.Pages = append(page.Site.Pages, listpage)
	listpage.peek()

	Paginators[listpath] = &Paginator{
		Number:      n,
		PathPattern: pathPattern,
		Page:        listpage,
		Pages:       make(PageSlice, 0),
	}
}

func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func ProcessPaginateCollectPages(page *Page, args []string) {
	length, err := strconv.Atoi(args[0])
	if err != nil {
		errexit(err)
	}

	pagi := Paginators[page.Source]
	paginated := Paginated[pagi.PathPattern]
	// NOTE: this hack for calling .Sort only once relies on the fact that
	// site.Pages are sorted by .ModTime (if they don't have .Date), and
	// .ModTime depends on a pagi.Number.
	if pagi.Number == 1 {
		paginated.Sort()
	}

	pagi.Pages = paginated[(pagi.Number-1)*length : MinInt(len(paginated), pagi.Number*length)]
}

//================ Pagination end

package processors

import (
	"errors"
	"fmt"
	gostatic "github.com/piranha/gostatic/lib"
	"strconv"
	"strings"
	"time"
)

type PaginateProcessor struct {
	collectPages bool
}

func NewPaginateProcessor() *PaginateProcessor {
	if _, ok := gostatic.TemplateFuncMap["paginator"]; !ok {
		gostatic.TemplateFuncMap["paginator"] = CurrentPaginator
	}
	return &PaginateProcessor{}
}

func NewPaginateCollectPagesProcessor() *PaginateProcessor {
	return &PaginateProcessor{true}
}

func (p *PaginateProcessor) Process(page *gostatic.Page, args []string) error {
	if !p.collectPages {
		return ProcessPaginate(page, args)
	}
	return ProcessPaginateCollectPages(page, args)
}

func (p *PaginateProcessor) Description() string {
	return "read config from content (separated by '----\\n')"
}

func (p *PaginateProcessor) Mode() int {
	if p.collectPages {
		return gostatic.Hidden
	}
	return gostatic.Pre
}

type Paginator struct {
	Number      int
	PathPattern string
	Page        *gostatic.Page
	Pages       gostatic.PageSlice
}

var Paginated = map[string]gostatic.PageSlice{}

var Paginators = map[string]*Paginator{}

func CurrentPaginator(current *gostatic.Page) *Paginator {
	// from processors.go
	return Paginators[current.Source]
}

func NewPaginator() *Paginator {
	p := &Paginator{
	//Paginated:  map[string]PageSlice{},
	//Paginators: map[string]*Paginator{},
	}
	return p
}

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

func ProcessPaginate(page *gostatic.Page, args []string) error {
	if len(args) < 2 {
		return errors.New("'paginate' rule needs two arguments")
	}
	length, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	pathPattern := args[1]

	if pages, ok := Paginated[pathPattern]; ok {
		Paginated[pathPattern] = append(pages, page)
	} else {
		Paginated[pathPattern] = gostatic.PageSlice{page}
	}

	site := page.Site

	// page number, 1-based
	n := 1 + ((len(Paginated[pathPattern]) - 1) / length)
	listpath := strings.Replace(pathPattern, "*", strconv.Itoa(n), 1)
	listpage := site.Pages.BySource(listpath)

	//todo catch this
	if listpage != nil {
		return nil
	}

	pattern, rules := site.Rules.MatchedRules(listpath)
	if rules == nil {
		return fmt.Errorf("Paginators path '%s' does not match any rule",
			listpath)
	}

	if len(rules) > 1 {
		return fmt.Errorf("Path '%s' matches multiple rules, pagination is not supported for this case",
			listpath)
	}

	rule := rules[0]

	if !strings.HasPrefix(string(rule.Commands[0]), "paginate-collect-pages") {
		rule.Commands = append(
			gostatic.CommandList{gostatic.Command("paginate-collect-pages " + args[0])},
			rule.Commands...)
	}

	listpage = &gostatic.Page{
		PageHeader: gostatic.PageHeader{Title: strconv.Itoa(n)},
		Site:       site,
		Pattern:    pattern,
		Rule:       rule,
		Source:     listpath,
		Path:       listpath,
		ModTime:    time.Unix(int64(n), 0),
	}
	listpage.SetWasRead(true)
	page.Site.Pages = append(page.Site.Pages, listpage)
	listpage.Peek()

	Paginators[listpath] = &Paginator{
		Number:      n,
		PathPattern: pathPattern,
		Page:        listpage,
		Pages:       make(gostatic.PageSlice, 0),
	}
	return nil
}

func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func ProcessPaginateCollectPages(page *gostatic.Page, args []string) error {
	length, err := strconv.Atoi(args[0])
	if err != nil {
		return err
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
	return nil
}

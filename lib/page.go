// (c) 2012 Alexander Solovyov
// under terms of ISC license

package gostatic

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	StateUnknown = iota
	StateChanged
	StateUnchanged
	StateIgnored
)

type Page struct {
	PageHeader

	Site    *Site `json:"-"`
	Rule    *Rule
	Pattern string
	Deps    PageSlice `json:"-"`

	Source  string
	Path    string
	ModTime time.Time

	processed bool
	state     int
	raw       string
	content   string
	wasread   bool // if content was read already
}

type PageSlice []*Page

func NewPages(site *Site, path string) PageSlice {
	stat, err := os.Stat(path)
	errhandle(err)

	relpath, err := filepath.Rel(site.Source, path)
	errhandle(err)

	// convert windows path separators to unix style
	relpath = strings.Replace(relpath, "\\", "/", -1)

	pattern, rules := site.Rules.MatchedRules(relpath)
	if rules == nil {
		rules = make([]*Rule, 1)
	}

	pages := make(PageSlice, 0)

	for _, rule := range rules {
		page := &Page{
			Site:    site,
			Rule:    rule,
			Pattern: pattern,
			Source:  relpath,
			Path:    relpath,
			ModTime: stat.ModTime(),
		}
		page.Peek()
		debug("Found page: %s; rule: %v\n",
			page.Source, page.Rule)
		pages = append(pages, page)
	}
	return pages
}

func (page *Page) Raw() string {
	if !page.wasread {
		data, err := ioutil.ReadFile(page.FullPath())
		errhandle(err)
		page.raw = string(data)
		page.wasread = true
		debug("Page '%s' was read, is of length %d\n", page.FullPath(), len(page.raw))
	}
	return page.raw
}

func (page *Page) Content() string {
	if page.content == "" {
		return page.Raw()
	}
	return page.content
}

func (page *Page) SetContent(content string) {
	page.content = content
}

func (page *Page) SetState(state int) {
	page.state = state
}

func (page *Page) FullPath() string {
	return filepath.Join(page.Site.Source, page.Source)
}

func (page *Page) OutputPath() string {
	return filepath.Join(page.Site.Output, page.Path)
}

func (page *Page) Url() string {
	if page == nil {
		errexit(fmt.Errorf(".Url called on a Page which does not exist"))
	}
	url := strings.Replace(page.Path, string(filepath.Separator), "/", -1)
	if url == "index.html" {
		return ""
	}
	if strings.HasSuffix(url, "/index.html") {
		return strings.TrimSuffix(url, "/index.html") + "/"
	}
	return url
}

func (page *Page) Name() string {
	return filepath.Base(page.Url())
}

func (page *Page) UrlTo(other *Page) string {
	return page.Rel(other.Url())
}

func (page *Page) Rel(path string) string {
	root := strings.Repeat("../", strings.Count(page.Url(), "/"))
	if root == "" {
		root = "./"
	}
	if path[0] == '/' {
		return root + path[1:]
	}
	return root + path
}

func (page *Page) Is(path string) bool {
	return page.Url() == path || page.Path == path
}

// Is used for dynamically created pages
func (page *Page) SetWasRead(wasread bool) {
	page.wasread = wasread
}

func (page *Page) WasRead() bool {
	return page.wasread
}

// Peek is used to run those processors which should be done before others can
// find out about us. Two actual examples include 'config' and 'rename'
// processors.
func (page *Page) Peek() error {
	if page.Rule == nil {
		return nil
	}

	for _, cmd := range page.Rule.Commands {
		err := page.Site.ProcessCommand(page, &cmd, true)
		if err != nil {
			return err
		}
	}

	// Raw is page content after preprocessors, but before preprocessors
	if page.content != "" {
		page.raw = page.content
	}
	return nil
}

func (page *Page) findDeps() {
	if page.Rule == nil {
		return
	}

	deps := make(PageSlice, 0)
	for _, other := range page.Site.Pages {
		if other != page && page.Rule.IsDep(other) {
			deps = append(deps, other)
		}
	}
	page.Deps = deps
}

func (page *Page) Changed() bool {
	if page.Site.ForceRefresh {
		return true
	}

	if page.state == StateUnknown {
		page.state = StateUnchanged
		dest, err := os.Stat(page.OutputPath())

		if err != nil ||
			dest.ModTime().Before(page.ModTime) ||
			dest.ModTime().Before(page.Site.ChangedAt) {
			page.state = StateChanged
		} else {
			for _, dep := range page.Deps {
				if dep.Changed() {
					page.state = StateChanged
				}
			}
		}
	}

	return page.state == StateChanged
}

func (page *Page) Process() (*Page, error) {
	if page.processed || page.Rule == nil {
		return page, nil
	}

	page.processed = true
	if page.Rule.Commands != nil {
		for _, cmd := range page.Rule.Commands {
			err := page.Site.ProcessCommand(page, &cmd, false)
			if err != nil {
				return nil, err
			}
		}
	}

	return page, nil
}

func (page *Page) WriteTo(writer io.Writer) (n int64, err error) {
	if page.Rule == nil {
		return 0, nil
	}

	if !page.processed {
		page.Process()
	}

	nint, err := writer.Write([]byte(page.Content()))
	return int64(nint), err
}

func (page *Page) Render() (n int64, err error) {
	if page.Rule == nil {
		return CopyFile(page.FullPath(), page.OutputPath())
	}

	file, err := os.Create(page.OutputPath())
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return page.WriteTo(file)
}

func (page *Page) UrlMatches(regex string) bool {
	re, err := regexp.Compile(regex)
	if err != nil {
		errhandle(fmt.Errorf("Incorrect regex given to Page.UrlMatches: '%s' ", regex))
	}
	return re.Match([]byte(page.Url()))
}

func (page *Page) Prev() *Page {
	return page.Site.Pages.Prev(page)
}

func (page *Page) Next() *Page {
	return page.Site.Pages.Next(page)
}

// PageSlice manipulation

func (pages PageSlice) Get(i int) *Page { return pages[i] }
func (pages PageSlice) First() *Page    { return pages.Get(0) }
func (pages PageSlice) Last() *Page     { return pages.Get(len(pages) - 1) }

func (pages PageSlice) Prev(cur *Page) *Page {
	for i, page := range pages {
		if page == cur {
			if i == pages.Len()-1 {
				return nil
			}
			return pages[i+1]
		}
	}
	return nil
}

func (pages PageSlice) Next(cur *Page) *Page {
	for i, page := range pages {
		if page == cur {
			if i == 0 {
				return nil
			}
			return pages[i-1]
		}
	}
	return nil
}

func (pages PageSlice) Slice(from int, to int) PageSlice {
	length := len(pages)

	if from > length {
		from = length
	}
	if to > length {
		to = length
	}

	return pages[from:to]
}

// Sorting interface
func (pages PageSlice) Len() int {
	return len(pages)
}

func (pages PageSlice) Less(i, j int) bool {
	left := pages.Get(i)
	right := pages.Get(j)
	if left.Date.Unix() == right.Date.Unix() {
		return left.Path > right.Path
	}
	return left.Date.Unix() > right.Date.Unix()
}

func (pages PageSlice) Swap(i, j int) {
	pages[i], pages[j] = pages[j], pages[i]
}

func (pages PageSlice) Sort() {
	sort.Sort(pages)
}

func (pages PageSlice) Reverse() {
	sort.Sort(sort.Reverse(pages))
}

func (pages PageSlice) Children(root string) *PageSlice {
	children := make(PageSlice, 0)

	for _, page := range pages {
		if !page.Hide &&
			strings.HasPrefix(page.Url(), root) &&
			page.Url() != root {
			children = append(children, page)
		}
	}

	return &children
}

func (pages PageSlice) WithTag(tag string) *PageSlice {
	tagged := make(PageSlice, 0)

	for _, page := range pages {
		if !page.Hide &&
			page.Tags != nil &&
			SliceStringIndexOf(page.Tags, tag) != -1 {
			tagged = append(tagged, page)
		}
	}

	return &tagged
}

func (pages PageSlice) HasPage(check func(page *Page) bool) bool {
	for _, page := range pages {
		if check(page) {
			return true
		}
	}
	return false
}

func (pages PageSlice) BySource(s string) *Page {
	for _, page := range pages {
		if page.Source == s {
			return page
		}
	}
	return nil
}

func (pages PageSlice) GlobSource(pattern string) *PageSlice {
	found := make(PageSlice, 0)

	for _, page := range pages {
		if matched, _ := path.Match(pattern, page.Source); matched {
			found = append(found, page)
		}
	}

	return &found
}

func (pages PageSlice) ByPath(s string) *Page {
	for _, page := range pages {
		if page.Path == s {
			return page
		}
	}
	return nil
}

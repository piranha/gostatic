package main

import (
	"fmt"
	"hash/adler32"
	"io"
	"regexp"
	"strings"
	"text/template"
)

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

func Truncate(length int, value string) string {
	if length > len(value) {
		length = len(value)
	}
	return value[0:length]
}

func StripHTML(value string) string {
	return regexp.MustCompile("<[^>]+>").ReplaceAllString(value, "")
}

func CurrentPaginator(current *Page) *Paginator {
	// from processors.go
	return Paginators[current.Source]
}

var TemplateFuncMap = template.FuncMap{
	"changed":    HasChanged,
	"cut":        Cut,
	"hash":       Hash,
	"version":    Versionize,
	"truncate":   Truncate,
	"strip_html": StripHTML,
	"split":      strings.Split,
	"paginator":  CurrentPaginator,
}

// (c) 2012 Alexander Solovyov
// under terms of ISC license

// Generating example site

package gostatic

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var ExampleConfig = `
TEMPLATES = site.tmpl
SOURCE = src
OUTPUT = site
TITLE = Example Site
URL = http://example.com/
AUTHOR = Your Name

blog/*.md:
	config
	ext .html
	directorify
	tags tags/*.tag
	markdown
	template post
	template page

*.tag: blog/*.md
	ext .html
	directorify
	template tag
	markdown
	template page

blog.atom: blog/*.md
	inner-template

index.html: blog/*.md
	config
	inner-template
	template page
`

var ExampleMakefile = `
GS ?= gostatic

compile:
	$(GS) config

w:
	$(GS) -w config
`

var ExampleTemplate = `
{{ define "header" }}<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="author" content="{{ html .Site.Other.Author }}">
  <link rel="alternate" type="application/atom+xml" title="{{ html .Site.Other.Title }} feed" href="{{ .Rel "blog.atom" }}">
  <title>{{ .Site.Other.Title }}{{ if .Title }}: {{ .Title }}{{ end }}</title>
  <link href="//netdna.bootstrapcdn.com/bootstrap/3.0.0/css/bootstrap.min.css" rel="stylesheet">
  <link rel="stylesheet" type="text/css" href="{{ .Rel "static/style.css" }}">
</head>
<body>
{{ end }}

{{ define "footer" }}
</body>
</html>
{{ end }}

{{define "date"}}
<time datetime="{{ .Format "2006-01-02T15:04:05Z07:00" }}">
  {{ .Format "2006, January 02" }}
</time>
{{end}}

{{ define "page" }}{{ template "header" . }}
  {{ .Content }}
{{ template "footer" . }}{{ end }}

{{ define "post" }}
<article>
  <header>
    <h1>{{ .Title }}</h1>
    <div class="info">
      {{ template "date" .Date }} &mdash;
      {{ range $i, $t := .Tags }}{{if $i}},{{end}}
      <a href="/tags/{{ $t }}/">{{ $t }}</a>{{ end }}
    </div>
  </header>
  <section>
  {{ .Content }}
  </section>
</article>
{{ end }}

{{define "tag"}}
# Pages tagged with {{ .Title }}
{{ range .Site.Pages.WithTag .Title }}
- [{{ .Title }}](../../{{ .Url }})
{{ end }}
{{ end }}
`

var ExampleIndex = `
title: Main Page
----
<ul class="post-list">
{{ range .Site.Pages.Children "blog/" }}
  <li>
    {{ template "date" .Date }} - <a href="{{ $.Rel .Url }}">{{ .Title }}</a>
  </li>
{{ end }}
</ul>
`

var ExampleFeed = `
<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:thr="http://purl.org/syndication/thread/1.0">
  <id>{{ .Site.Other.Url }}</id>
  <title>{{ .Site.Other.Title }}</title>
  {{ with .Site.Pages.Children "blog/" }}
  <updated>{{ .First.Date.Format "2006-01-02T15:04:05Z07:00" }}</updated>
  {{ end }}
  <author><name>{{ .Site.Other.Author }}</name></author>
  <link href="{{ .Site.Other.Url }}" rel="alternate"></link>
  <generator uri="http://github.com/piranha/gostatic/">gostatic</generator>

{{ with .Site.Pages.Children "blog/" }}
{{ range .Slice 0 5 }}
<entry>
  <id>{{ .Url }}</id>
  <author><name>{{ or .Other.Author .Site.Other.Author }}</name></author>
  <title type="html">{{ html .Title }}</title>
  <published>{{ .Date.Format "2006-01-02T15:04:05Z07:00" }}</published>
  {{ range .Tags }}
  <category term="{{ . }}"></category>
  {{ end }}
  <link href="{{ .Site.Other.Url }}/{{ .Url }}" rel="alternate"></link>
  <content type="html">
    {{/* .Process runs here in case only feed changed */}}
    {{ with cut "<section>" "</section>" .Process.Content }}
      {{ html . }}
    {{ end }}
  </content>
</entry>
{{ end }}
{{ end }}
</feed>
`

var ExamplePost = `
title: First Post
date: 2012-12-12
tags: blog
----
My first post with [gostatic](http://github.com/piranha/gostatic/).
`

var ExampleStyle = `
/* put your style rules here */
`

func WriteExample(dir string) error {
	WriteFile(dir, "config", ExampleConfig)
	WriteFile(dir, "Makefile", ExampleMakefile)
	WriteFile(dir, "site.tmpl", ExampleTemplate)
	WriteFile(dir, "src/index.html", ExampleIndex)
	WriteFile(dir, "src/blog/first.md", ExamplePost)
	WriteFile(dir, "src/blog.atom", ExampleFeed)
	WriteFile(dir, "src/static/style.css", ExampleStyle)
	return nil
}

func WriteFile(dir string, fn string, content string) error {
	path := filepath.Join(dir, filepath.FromSlash(fn))
	dir = filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	towrite := []byte(strings.TrimLeftFunc(content, unicode.IsSpace))
	if err := ioutil.WriteFile(path, towrite, 0644); err != nil {
		return err
	}

	return nil
}

package main

import (
	"os"
	"time"
	"bytes"
	"io/ioutil"
	"text/template"
	blackfriday "github.com/russross/blackfriday"
)

type Page struct {
	Config
	Content []byte
	Path string
	Mod time.Time
	Template *template.Template
}

func NewPage(path string, base string, t *template.Template) *Page {
	text, err := ioutil.ReadFile(path)
	errhandle(err)

	stat, err := os.Stat(path)
	errhandle(err)

	head, content := SplitHead(text)

	return &Page{
		Config: *ParseConfig(string(head)),
		Content: blackfriday.MarkdownCommon(content),
		Path: path,
		Mod: stat.ModTime(),
		Template: t,
	}
}

func SplitHead(text []byte) ([]byte, []byte) {
	parts := bytes.SplitN(text, []byte("----\n"), 2)
	return parts[0], parts[1]
}

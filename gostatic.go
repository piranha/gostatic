// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"fmt"
	goopt "github.com/droundy/goopt"
	"text/template"
)

var Version = "0.1"

var Summary = `gostatic -t template [-t template] sitedir

Build a site
`

var templates = goopt.Strings([]string{"-t", "--template"},
	"template", "path to template")
var output = goopt.String([]string{"-o", "--output"},
	"directory", "output directory")
var showVersion = goopt.Flag([]string{"-v", "--version"}, []string{},
	"show version and exit", "")

func main() {
	goopt.Version = Version
	goopt.Summary = Summary

	goopt.Parse(nil)

	if *showVersion {
		fmt.Printf("gostatic %s\n", goopt.Version)
		return
	}

	if len(*templates) == 0 || len(goopt.Args) == 0 {
		println(goopt.Usage())
		return
	}

	t, err := template.ParseFiles(*templates...)
	errhandle(err)

	site := NewSite(t, goopt.Args[0])
	site.Summary()
}

// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"encoding/json"
	"fmt"
	goopt "github.com/droundy/goopt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var Version = "0.1"

var Summary = `gostatic [OPTIONS] path/to/config

Build a site.
`

var showProcessors = goopt.Flag([]string{"--processors"}, []string{},
	"show internal processors", "")
var showConfig = goopt.Flag([]string{"--show-config"}, []string{},
	"dump config as JSON on stdout", "")
var showSummary = goopt.Flag([]string{"--summary"}, []string{},
	"print everything on stdout", "")
var initExample = goopt.Flag([]string{"--init"}, []string{},
	"init example site", "")

// used in Page.Changed()
var force = goopt.Flag([]string{"-f", "--force"}, []string{},
	"force building all pages", "")

var doWatch = goopt.Flag([]string{"-w", "--watch"}, []string{},
	"watch for changes and serve them as http", "")
var port = goopt.String([]string{"-p", "--port"}, "8000",
	"port to serve on")

var verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{},
	"enable verbose output", "")
var showVersion = goopt.Flag([]string{"-V", "--version"}, []string{},
	"show version and exit", "")

func main() {
	goopt.Version = Version
	goopt.Summary = Summary

	goopt.Parse(nil)

	if *showSummary && *doWatch {
		errhandle(fmt.Errorf("--summary and --watch do not mix together well"))
	}

	if *showVersion {
		out("gostatic %s\n", goopt.Version)
		return
	}

	if *showProcessors {
		InitProcessors()
		ProcessorSummary()
		return
	}

	if *initExample {
		cwd, _ := os.Getwd()
		WriteExample(cwd)
		return
	}

	if len(goopt.Args) == 0 {
		println(goopt.Usage())
		return
	}

	InitProcessors()
	config, err := NewSiteConfig(goopt.Args[0])
	errhandle(err)

	if *showConfig {
		x, err := json.MarshalIndent(config, "", "  ")
		errhandle(err)
		println(string(x))
		return
	}

	site := NewSite(config)
	if *showSummary {
		site.Summary()
	} else {
		site.Render()
	}

	if *doWatch {
		StartWatcher(config)
		out("Starting server at *:%s...\n", *port)

		fs := http.FileServer(http.Dir(config.Output))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fs.ServeHTTP(w, r)
		})

		err := http.ListenAndServe(":"+*port, nil)
		errhandle(err)
	}
}

func StartWatcher(config *SiteConfig) {
	filemod, err := DirWatcher(config.Source)
	errhandle(err)

	go func() {
		for {
			fn := <-filemod
			if !strings.HasPrefix(filepath.Base(fn), ".") {
				site := NewSite(config)
				site.Render()
			}
		}
	}()
}

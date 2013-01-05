// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"encoding/json"
	"fmt"
	goopt "github.com/droundy/goopt"
	"net/http"
	"path/filepath"
	"strings"
)

var Version = "0.1"

var Summary = `gostatic path/to/config

Build a site.
`

var showVersion = goopt.Flag([]string{"-V", "--version"}, []string{},
	"show version and exit", "")
var showProcessors = goopt.Flag([]string{"--processors"}, []string{},
	"show internal processors", "")
var showSummary = goopt.Flag([]string{"--summary"}, []string{},
	"print everything on stdout", "")
var showConfig = goopt.Flag([]string{"--show-config"}, []string{},
	"dump config as JSON on stdout", "")
var doWatch = goopt.Flag([]string{"-w", "--watch"}, []string{},
	"watch for changes and serve them as http", "")
var port = goopt.String([]string{"-p", "--port"}, "8000",
	"port to serve on")
var verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{},
	"enable verbose output", "")

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
		ProcessorSummary()
		return
	}

	if len(goopt.Args) == 0 {
		println(goopt.Usage())
		return
	}

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
		});

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

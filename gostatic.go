// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"encoding/json"
	"fmt"
	goopt "github.com/droundy/goopt"
	"io/ioutil"
	"path/filepath"
	"github.com/howeyc/fsnotify"
	"strings"
	"net/http"
)

var Version = "0.1"

var Summary = `gostatic path/to/config.json

Build a site.
`

var showVersion = goopt.Flag([]string{"-v", "--version"}, []string{},
	"show version and exit", "")
var showProcessors = goopt.Flag([]string{"--processors"}, []string{},
	"show internal processors", "")
var showSummary = goopt.Flag([]string{"--summary"}, []string{},
	"print everything on stdout", "")
var doWatch = goopt.Flag([]string{"-w", "--watch"}, []string{},
	"watch for changes and serve them as http", "")
var port = goopt.String([]string{"-p", "--port"}, "8000",
	"port to serve on")

type GlobalConfig struct {
	Templates []string
	Source    string
	Output    string
	Rules     RuleMap
}

func main() {
	goopt.Version = Version
	goopt.Summary = Summary

	goopt.Parse(nil)

	if *showSummary && *doWatch {
		errhandle(fmt.Errorf("--summary and --watch do not mix together well"))
	}

	if *showVersion {
		fmt.Printf("gostatic %s\n", goopt.Version)
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

	text, err := ioutil.ReadFile(goopt.Args[0])
	errhandle(err)
	cfg, err := NewSiteConfig(string(text))
	errhandle(err)
	x, err := json.Marshal(cfg)
	errhandle(err)
	println(string(x))
	return

	config := RetrieveGlobalConfig(goopt.Args[0])

	site := NewSite(config)
	if *showSummary {
		site.Summary()
	} else {
		site.Render()
	}

	if *doWatch {
		StartWatcher(config)
		fmt.Printf("Starting server at *:%s...\n", *port)
		err := http.ListenAndServe(":" + *port,
			http.FileServer(http.Dir(config.Output)))
		errhandle(err)
	}
}

func RetrieveGlobalConfig(path string) *GlobalConfig {
	conftext, err := ioutil.ReadFile(path)
	errhandle(err)

	var config GlobalConfig
	err = json.Unmarshal(conftext, &config)
	errhandle(err)

	basepath, _ := filepath.Split(path)
	config.Source = filepath.Join(basepath, config.Source)
	config.Output = filepath.Join(basepath, config.Output)

	templates := make([]string, len(config.Templates))
	for i, template := range config.Templates {
		templates[i] = filepath.Join(basepath, template)
	}
	config.Templates = templates

	return &config
}

func StartWatcher(config *GlobalConfig) {
	watcher, err := fsnotify.NewWatcher()
	errhandle(err)

	go func() {
		for {
			select {
			case ev := <- watcher.Event:
				if !strings.HasPrefix(filepath.Base(ev.Name), ".") {
					site := NewSite(config)
					site.Render()
				}
			case err := <- watcher.Error:
				errhandle(err)
			}
		}
	}()

	err = watcher.Watch(config.Source)
	errhandle(err)
}

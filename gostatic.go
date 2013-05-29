// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"encoding/json"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var Version = "1.1"

var opts struct {
	ShowProcessors bool    `long:"processors" description:"show page processors"`
	ShowConfig     bool    `long:"show-config" description:"print config as JSON"`
	ShowSummary    bool    `long:"summary" description:"print all pages on stdout"`
	InitExample    *string `short:"i" long:"init" description:"create example site"`

	// used in Page.Changed()
	Force bool `short:"f" long:"force" description:"force building all pages"`

	Watch bool   `short:"w" long:"watch" description:"serve site on HTTP and rebuild on changes"`
	Port  string `short:"p" long:"port" default:"8000" description:"port to serve on"`

	Verbose bool `short:"v" long:"verbose" description:"enable verbose output"`
	Version bool `short:"V" long:"version" description:"show version and exit"`
}

func main() {
	argparser := flags.NewParser(&opts,
		flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	argparser.Usage = "[OPTIONS] path/to/config\n\nBuild a site."

	args, err := argparser.Parse()
	if err != nil {
		return
	}

	if opts.ShowSummary && opts.Watch {
		errhandle(fmt.Errorf("--summary and --watch do not mix together well"))
	}

	if opts.Version {
		out("gostatic %s\n", Version)
		return
	}

	if opts.ShowProcessors {
		InitProcessors()
		ProcessorSummary()
		return
	}

	if opts.InitExample != nil {
		target, _ := os.Getwd()
		if len(*opts.InitExample) > 0 {
			target = filepath.Join(target, *opts.InitExample)
		}
		WriteExample(target)
		return
	}

	if len(args) == 0 {
		argparser.WriteHelp(os.Stdout)
		return
	}

	InitProcessors()
	config, err := NewSiteConfig(args[0])
	errhandle(err)

	if opts.ShowConfig {
		x, err := json.MarshalIndent(config, "", "  ")
		errhandle(err)
		println(string(x))
		return
	}

	site := NewSite(config)
	if opts.ShowSummary {
		site.Summary()
	} else {
		site.Render()
	}

	if opts.Watch {
		StartWatcher(config)
		out("Starting server at *:%s...\n", opts.Port)

		fs := http.FileServer(http.Dir(config.Output))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fs.ServeHTTP(w, r)
		})

		err := http.ListenAndServe(":"+opts.Port, nil)
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

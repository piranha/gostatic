// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"gopkg.in/fsnotify.v1"
	"os"
	"path/filepath"
)

func Watcher(config *SiteConfig) (chan string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	evs := make(chan string, 10)

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if ev.Op&fsnotify.Create == fsnotify.Create {
					watcher.Add(ev.Name)
				} else if ev.Op&fsnotify.Remove == fsnotify.Remove {
					watcher.Remove(ev.Name)
				}

				evs <- ev.Name
			case err := <-watcher.Errors:
				errhandle(err)
			}
		}
	}()

	filepath.Walk(config.Source, watchAll(watcher))
	for _, path := range config.Templates {
		watcher.Add(path)
	}

	return evs, nil
}

func watchAll(watcher *fsnotify.Watcher) filepath.WalkFunc {
	return func(fn string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		watcher.Add(fn)
		return nil
	}
}

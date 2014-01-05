// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path/filepath"
)

func Watcher(config *SiteConfig) (chan string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 10)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsCreate() {
					watcher.Watch(ev.Name)
				} else if ev.IsDelete() {
					watcher.RemoveWatch(ev.Name)
				}
				ch <- ev.Name
			}
		}
	}()

	filepath.Walk(config.Output, watchAll(watcher))
	for _, path := range config.Templates {
		watcher.Watch(path)
	}

	return ch, nil
}

func watchAll(watcher *fsnotify.Watcher) filepath.WalkFunc {
	return func(fn string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		watcher.Watch(fn)
		return nil
	}
}

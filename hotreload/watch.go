// (c) 2012 Alexander Solovyov
// under terms of ISC license

package hotreload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func watchAll(watcher *fsnotify.Watcher) filepath.WalkFunc {
	return func(fn string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		watcher.Add(fn)
		return nil
	}
}

func fileWatcher(dirs, files []string) (chan string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	evs := make(chan string, 50)

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if strings.HasPrefix(filepath.Base(ev.Name), ".") {
					continue
				}

				if ev.Op&fsnotify.Create == fsnotify.Create {
					watcher.Add(ev.Name)
				} else if ev.Op&fsnotify.Remove == fsnotify.Remove {
					watcher.Remove(ev.Name)
				}

				evs <- ev.Name
			case err := <-watcher.Errors:
				fmt.Printf("Error: %s\n", err)
			}
		}
	}()

	for _, path := range dirs {
		filepath.Walk(path, watchAll(watcher))
	}
	for _, path := range files {
		watcher.Add(path)
	}

	return evs, nil
}

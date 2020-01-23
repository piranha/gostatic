// This functionality is made by looking at source of devd:
// https://github.com/cortesi/devd/blob/master/livereload/livereload.go
// which is copyrighted under MIT by Aldo Cortesi

package hotreload

import (
	"net/http"
	"time"

	//	"github.com/gorilla/websocket"
)

const (
	// path to websocket endpoint
	EndpointPath = "/.gostatic.live"
	// path to livereload support js
	LiveScriptPath = "/.gostatic.live.js"
)



// ServeHTTP serves files from `dir` and adds livereload support to html files
func ServeHTTP(dir, port string) error {
	fs := http.FileServer(http.Dir(dir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		fs.ServeHTTP(w, r)
	})

	return http.ListenAndServe(":"+port, nil)
}

// Watch starts file watcher
func Watch(dirs, files []string, callback func()) error {
	filemods, err := fileWatcher(dirs, files)
	if err != nil {
		return err
	}

	go func() {
		for {
			<-filemods
			time.Sleep(10 * time.Millisecond)
			for len(filemods) > 0 {
				<-filemods
			}
			callback()
		}
	}()

	return nil
}

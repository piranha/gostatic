// This functionality is made by looking at source of devd:
// https://github.com/cortesi/devd/blob/master/livereload/livereload.go
// which is copyrighted under MIT by Aldo Cortesi

package hotreload

import (
	"net/http"
	"time"
	"sync"
	"fmt"
	"strings"
	"regexp"

	"github.com/gorilla/websocket"
)

const (
	// EndpointPath is a path to websocket endpoint
	EndpointPath = "/.gostatic.hotreload"
	// ScriptPath is a path to hotreload support js
	ScriptPath = "/.gostatic.hotreload.js"
)

type Server struct {
	sync.Mutex
	conns     map[*websocket.Conn]bool

	filemods <-chan string
	dirs     []string
	files    []string
}

func NewServer(path string) (*Server, error) {
	filemods, err := fileWatcher([]string{path}, []string{})
	if err != nil {
		return nil, err
	}

	s := &Server{
		conns: make(map[*websocket.Conn]bool),
		filemods: filemods,
	}
	go s.run()
	return s, nil
}

func (s *Server) run() {
	for {
		fns := []string{}
		fns = append(fns, <-s.filemods)
		time.Sleep(16 * time.Millisecond)
		for len(s.filemods) > 0 {
			fns = append(fns, <-s.filemods)
		}

		cmd := "css"
		for _, fn := range fns {
			if !strings.HasSuffix(fn, ".css") {
				cmd = "page"
			}
		}

		s.Lock()
		for conn := range s.conns {
			if conn == nil {
				continue
			}
			err := conn.WriteMessage(websocket.TextMessage, []byte(cmd))
			if err != nil {
				fmt.Errorf("error: %s", err)
				delete(s.conns, conn)
			}
		}
		s.Unlock()
	}

	s.Lock()
	for conn := range s.conns {
		delete(s.conns, conn)
		conn.Close()
	}
	s.Unlock()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (s *Server) Upgrade(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Errorf("can't establish WebSocket connection: %s", err)
		http.Error(w, "Can't upgrade.", 500)
		return
	}
	s.Lock()
	s.conns[conn] = true
	s.Unlock()
}


type Injector struct {
	http.ResponseWriter
}

var htmlRe = regexp.MustCompile("\btext/html\b")
var injectorRe = regexp.MustCompile(`<\/head>`)
var payload = []byte(`<script src="/.gostatic.hotreload.js" async></script></head>`)


func (inj *Injector) WriteHeader(statusCode int) {
	inj.ResponseWriter.Header().Del("Content-Length")
	inj.ResponseWriter.WriteHeader(statusCode)
}

func (inj *Injector) Write(data []byte) (int, error) {
	ctypes, ok := inj.ResponseWriter.Header()["Content-Type"]
	isHTML := false
	if ok {
		for _, ctype := range ctypes {
			if strings.HasPrefix(ctype, "text/html") {
				isHTML = true
			}
		}
	}

	if isHTML {
		data = injectorRe.ReplaceAllLiteral(data, payload)
	}
	return inj.ResponseWriter.Write(data)
}


// ServeHTTP serves files from `dir` and adds hotreload support to html files
func ServeHTTP(source, port string, hotreload bool) error {

	server, err := NewServer(source)
	if err != nil {
		return err
	}

	http.HandleFunc(EndpointPath, server.Upgrade)

	http.HandleFunc(ScriptPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(Morphdom)
		w.Write(Script)
	})

	fs := http.FileServer(http.Dir(source))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		if hotreload {
			injector := Injector{w}
			fs.ServeHTTP(&injector, r)
		} else {
			fs.ServeHTTP(w, r)
		}
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

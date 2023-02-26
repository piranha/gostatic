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
	"bytes"
	"io/ioutil"
	"os"

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


/// This code is surely full of pain, but I tried overriding WriteResponse and
/// it never worked properly :(

var (
	Head = []byte(`</head>`)
	Entry = []byte(`<script src="/.gostatic.hotreload.js" async></script></head>`)
)

type injectingFileInfo struct {
    os.FileInfo
    size int64
}

func (fi injectingFileInfo) Size() int64 {
    return fi.size
}


type injectingFile struct {
    http.File
    buffer *bytes.Buffer
}

func (f *injectingFile) Read(p []byte) (int, error) {
	if f.buffer == nil {
		content, err := ioutil.ReadAll(f.File)
		if err != nil {
			return 0, err
		}

		content = bytes.Replace(content, Head, Entry, 1)
		f.buffer = bytes.NewBuffer(content)
	}

    return f.buffer.Read(p)
}

func (f injectingFile) Stat() (os.FileInfo, error) {
    info, err := f.File.Stat()
    if err != nil {
        return nil, err
    }

    size := info.Size() + int64(len(Entry)-len(Head))
    modifiedInfo := &injectingFileInfo{info, size}

    return modifiedInfo, nil
}


type injectingFS struct {
	http.FileSystem
}

func (fsys injectingFS) Open(name string) (http.File, error) {
	file, err := fsys.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	if (strings.HasSuffix(name, ".html")) {
		return &injectingFile{file, nil}, nil
	}
	return file, err
}

/// End of injection madness


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

	var fs http.Handler
	if hotreload {
		fs = http.FileServer(injectingFS{http.Dir(source)})
	} else {
		fs = http.FileServer(http.Dir(source))
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

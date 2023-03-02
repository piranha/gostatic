// This functionality was first made by looking at source of devd:
// https://github.com/cortesi/devd/blob/master/livereload/livereload.go
// which is copyrighted under MIT by Aldo Cortesi

package hotreload

import (
	"net/http"
	"time"
	"fmt"
	"strings"
	"bytes"
	"io/ioutil"
	"os"
	// "log"
)

const (
	// EndpointPath is a path to websocket endpoint
	EndpointPath = "/.gostatic.hotreload"
	// ScriptPath is a path to hotreload support js
	ScriptPath = "/.gostatic.hotreload.js"
)

// Server-Sent Events

type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
    Notifier chan []byte

    // New client connections
    newConns chan chan []byte

    // Closed client connections
    closingConns chan chan []byte

    // Client connections registry
    conns map[chan []byte]bool
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	fmt.Fprintf(rw, "data: start\n\n")
	flusher.Flush()

	messageChan := make(chan []byte)
	broker.newConns <- messageChan
	defer func() {
		broker.closingConns <- messageChan
	}()

	// unregister client when connection closes
	notify := rw.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		broker.closingConns <- messageChan
	}()

	for {
		fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)
		flusher.Flush()
	}
}


func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newConns:
			broker.conns[s] = true
			// log.Printf("Conn added. %d registered conns", len(broker.conns))

		case s := <-broker.closingConns:
			delete(broker.conns, s)
			// log.Printf("Conn removed. %d registered conns", len(broker.conns))

		case event := <-broker.Notifier:
			for conn, _ := range broker.conns {
				conn <- event
			}
		}
	}
}


func NewServer() (broker *Broker) {
  broker = &Broker{
    Notifier:     make(chan []byte, 1),
    newConns:     make(chan chan []byte),
    closingConns: make(chan chan []byte),
    conns:        make(map[chan []byte]bool),
  }

  go broker.listen()
  return
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
	// server, err := NewServer(source)
	// if err != nil {
	// 	return err
	// }

	// http.HandleFunc(EndpointPath, server.Upgrade)
	filemods, err := fileWatcher([]string{source}, []string{})
	if err != nil {
		return err
	}

	ssebroker := NewServer()
	go func() {
		for {
			fns := []string{}
			fns = append(fns, <-filemods)
			time.Sleep(16 * time.Millisecond)
			for len(filemods) > 0 {
				fns = append(fns, <-filemods)
			}
			cmd := "css"
			for _, fn := range fns {
				if !strings.HasSuffix(fn, ".css") {
					cmd = "page"
				}
			}
			ssebroker.Notifier <- []byte(cmd)
		}
	}()

	http.HandleFunc(EndpointPath, ssebroker.ServeHTTP)

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

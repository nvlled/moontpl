package moontpl

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

const reloadFilename = "/.modified"

type WatcherFn func(string)

type FsWatcher struct {
	watcher   *fsnotify.Watcher
	listeners map[int]WatcherFn
	nextID    int

	// just use a simple mutex since
	// this will only be used on a local development server,
	// so performance doesn't really matter
	mu sync.Mutex
}

func newFsWatcher() *FsWatcher {
	return &FsWatcher{
		listeners: map[int]WatcherFn{},
		nextID:    1,
	}
}

func (fw *FsWatcher) Emit(filename string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	log.Printf("emit %v to %v listeners", filename, len(fw.listeners))
	for _, fn := range fw.listeners {
		fn(filename)
	}
}

func (fw *FsWatcher) On(fn WatcherFn) int {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	id := fw.nextID
	fw.nextID++
	fw.listeners[id] = fn

	return id
}

func (fw *FsWatcher) Off(id int) {
	fw.mu.Lock()
	delete(fw.listeners, id)
	fw.mu.Unlock()
}

func (m *Moontpl) StartFsWatch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	m.fsWatcher.watcher = watcher

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("file watcher stopped(?)")
					return
				}
				if event.Has(fsnotify.Write | fsnotify.Create) {
					log.Println("modified file:", event.Name)
					m.fsWatcher.Emit(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = fs.WalkDir(os.DirFS(m.SiteDir), ".", func(p string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir.IsDir() {
			filename := filepath.Join(m.SiteDir, p)
			log.Print("watching dir: ", filename)
			if err = watcher.Add(filename); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	<-make(chan struct{})

	return nil
}

func (m *Moontpl) handleCheckModified(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fsID := m.fsWatcher.On(func(link string) {
		resp := "event: message\ndata: _\n\n"
		_, err := w.Write([]byte(resp))
		if err != nil {
			log.Print(err)
		} else {
			w.(http.Flusher).Flush()
		}
	})

	<-r.Context().Done()
	m.fsWatcher.Off(fsID)

}

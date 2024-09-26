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
	watcher      *fsnotify.Watcher
	listeners    map[int]WatcherFn
	filesToWatch []string
	nextID       int

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

func (fw *FsWatcher) Add(filename string) {
	fw.filesToWatch = append(fw.filesToWatch, mustAbs(filename))
}

func (m *Moontpl) stopFsWatch() error {
	w := m.fsWatcher.watcher
	if w != nil {
		return w.Close()
	}
	return nil
}

func (m *Moontpl) startFsWatch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
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
		if p != "." && p[0] == '.' {
			return nil
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

	if fsExists("./lua") {
		dir := mustAbs("./lua")
		log.Print("watching dir: ", dir)
		if err := watcher.Add(dir); err != nil {
			panic(err)
		}
	}
	for _, filename := range m.fsWatcher.filesToWatch {
		log.Print("watching: ", filename)
		if err := watcher.Add(filename); err != nil {
			panic(err)
		}
	}

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

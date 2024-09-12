package moontpl

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func (m *Moontpl) Serve(addr string) {
	server := &http.Server{
		Addr:    addr,
		Handler: m.createHTTPHandler(),
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (m *Moontpl) createHTTPHandler() http.Handler {
	pageDir := http.FileServer(http.Dir(m.SiteDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pagePath := path.Clean(r.URL.Path)

		var filename string
		if pagePath == "/" {
			filename = path.Join(m.SiteDir, "index.html.lua")
		} else {
			filename = path.Join(m.SiteDir, pagePath)
		}

		stat, err := fsStat(filename)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			respondInternalError(w, err)
			return
		}

		if stat != nil && stat.IsDir() {
			filename = path.Join(filename, "index.html.lua")
		} else if !strings.HasSuffix(filename, ".lua") {
			filename += ".lua"
		}

		if !fsExists(filename) {
			r.URL.Path = path.Join("/", r.URL.Path)
			log.Println("serve file:", r.URL.Path)
			pageDir.ServeHTTP(w, r)
			return
		}

		log.Println("render lua:", filename)

		output, err := m.RenderFile(filename)
		if err != nil {
			respondInternalError(w, err)
			return
		}

		w.Write([]byte(output))
	})
}

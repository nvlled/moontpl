package moontpl

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
)

func (m *Moontpl) Serve(addr string) {
	server := &http.Server{
		Addr:    addr,
		Handler: m.createHTTPHandler(),
	}

	log.Printf("server listening at http://%s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (m *Moontpl) createHTTPHandler() http.Handler {
	pageDir := http.FileServer(http.Dir(m.SiteDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pagePath := path.Clean(r.URL.Path)

		if pagePath == reloadFilename {
			m.handleCheckModified(w, r)
			return
		}

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

		if filepath.Ext(filename) == ".lua" && !fsExists(filename) && !hasPathParams(filename) {
			r.URL.Path = path.Join("/", r.URL.Path)
			log.Println("serve file:", r.URL.Path)
			pageDir.ServeHTTP(w, r)
			return
		}

		log.Println("run file:", filename)
		output, err := m.RenderFile(filename)
		if err != nil {
			respondInternalError(w, err)
			return
		}

		ext := apply2(strings.TrimSuffix(filename, ".lua"), filepath.Ext, mime.TypeByExtension)

		w.Header().Add("Content-Type", ext)
		_, err = w.Write([]byte(output))
		if err != nil {
			respondInternalError(w, err)
		}
	})
}

func respondInternalError(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<pre>
%v
</pre>
<style>
pre {
	white-space: pre-wrap;
}         
body {
	font-size: 18px;
	color: #f00;
	background: #222;
}
</style>
<script>
	window.addEventListener("load", function() {
		new EventSource('/.modified').onmessage = function(event) {
			window.location.reload();
		};
	});
</script>
</html>
	`, err)
	log.Print(err)
	debug.PrintStack()
}

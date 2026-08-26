package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"
)

//go:embed dist
var Assets embed.FS

// GetHandler serves the new SPA (built by Vite into web/dist). Unlike the old
// UI (hash routing), this app uses history-mode routing, so any GET that does
// not resolve to an existing file falls back to index.html.
func GetHandler(useOS bool) http.Handler {
	var fsys http.FileSystem
	if useOS {
		fsys = http.Dir("web/dist")
	} else {
		sub, err := fs.Sub(Assets, "dist")
		if err != nil {
			log.Panic(err)
		}
		fsys = http.FS(sub)
	}
	fileServer := http.FileServer(fsys)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			fileServer.ServeHTTP(w, r)
			return
		}
		// If the request path maps to a real file (asset), serve it; otherwise
		// hand the route to the SPA.
		p := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if p == "" || p == "index.html" {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		if f, err := fsys.Open(p); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

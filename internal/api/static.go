package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func staticHandler(distFS fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean(r.URL.Path)
		if clean == "/" || clean == "." {
			serveIndex(distFS, w, r)
			return
		}
		upath := strings.TrimPrefix(clean, "/")
		f, err := distFS.Open(upath)
		if err != nil {
			serveIndex(distFS, w, r)
			return
		}
		stat, statErr := f.Stat()
		f.Close()
		if statErr != nil || stat.IsDir() {
			serveIndex(distFS, w, r)
			return
		}
		if strings.HasPrefix(clean, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.ServeFileFS(w, r, distFS, upath)
	})
}

func serveIndex(distFS fs.FS, w http.ResponseWriter, r *http.Request) {
	f, err := distFS.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	f.Close()
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFileFS(w, r, distFS, "index.html")
}

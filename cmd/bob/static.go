package main

import (
	"io/fs"
	"net/http"
)

// noDirFS hides directories from http.FileServer, so a request for one is a
// 404 rather than a directory listing.
type noDirFS struct {
	http.FileSystem
}

func (fsys noDirFS) Open(name string) (http.File, error) {
	f, err := fsys.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if info.IsDir() {
		f.Close()
		return nil, fs.ErrNotExist
	}
	return f, nil
}

// cacheStatic sets Cache-Control on static files. A URL carrying this
// release's ?v=assetVersion never changes, so browsers may keep it for a year.
// Anything else, including a ?v= from another release during a rolling
// deploy, must be revalidated.
func cacheStatic(assetVersion string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.URL.Query().Get("v"); v != "" && v == assetVersion {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

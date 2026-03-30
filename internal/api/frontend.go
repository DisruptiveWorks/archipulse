package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// serveFrontend mounts the embedded SPA onto the router.
// Assets under /assets/* are served directly; every other unmatched
// path returns index.html so the hash-based JS router handles navigation.
func serveFrontend(r *chi.Mux, static embed.FS) {
	sub, err := fs.Sub(static, "ui/dist")
	if err != nil {
		panic("frontend: sub fs: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	r.Handle("/assets/*", fileServer)

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFileFS(w, req, sub, "index.html")
	})
}

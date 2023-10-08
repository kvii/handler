// Package handler define custom http handlers.
package handler

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
)

// VueServer create a http.Handler to serve vue dist.
// You can use it like this:
//
//	http.Handle("/", handler.VueServer(http.Dir("dist")))
//
// It is useful when your project have HTML5 history mode.
//
// https://router.vuejs.org/guide/essentials/history-mode.html#HTML5-Mode
func VueServer(fs http.FileSystem) http.Handler {
	return &vue{
		h:  http.FileServer(fs),
		fs: fs,
	}
}

type vue struct {
	h  http.Handler
	fs http.FileSystem
}

func (v *vue) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f, err := v.fs.Open(path.Clean(r.URL.Path)); err == nil {
		f.Close()
		v.h.ServeHTTP(w, r)
		return
	}

	f, err := v.fs.Open("index.html")
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	http.ServeContent(w, r, "name", d.ModTime(), f)
}

// copy from http.toHTTPError
func toHTTPError(err error) (msg string, httpStatus int) {
	if errors.Is(err, fs.ErrNotExist) {
		return "404 page not found", http.StatusNotFound
	}
	if errors.Is(err, fs.ErrPermission) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

package serve

import (
	"io"
	"net/http"
	"time"
	_ "unsafe"
)

//go:linkname toHTTPError net/http.toHTTPError
//go:nosplit
func toHTTPError(error) (string, int)

//go:linkname serveContent net/http.serveContent
//go:nosplit
func serveContent(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	modtime time.Time,
	sizeFunc func() (int64, error),
	content io.ReadSeeker,
)

// ServeFile handles static file response.
func ServeFile(w http.ResponseWriter, r *http.Request, f http.File) {
	stat, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	} else if stat.IsDir() {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// serveContent will check modification time
	sizeFunc := func() (int64, error) { return stat.Size(), nil }
	serveContent(w, r, stat.Name(), stat.ModTime(), sizeFunc, f)
}

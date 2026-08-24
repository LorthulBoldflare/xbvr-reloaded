package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// serveFileFromDir streams the file at r.URL.Path from within baseDir.
// The cleaned absolute path must stay inside baseDir, otherwise the request
// is rejected — this blocks path traversal attempts (e.g. /download/../../etc/passwd).
// If attachment is true, a Content-Disposition header forces a browser download.
func serveFileFromDir(w http.ResponseWriter, r *http.Request, baseDir string, attachment bool) {
	path := filepath.Join(baseDir, r.URL.Path)

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		http.Error(w, "invalid base directory", http.StatusInternalServerError)
		return
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil || !strings.HasPrefix(pathAbs, baseAbs+string(os.PathSeparator)) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	fi, err := os.Stat(pathAbs)
	if err != nil || fi.IsDir() {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	reader, err := os.Open(pathAbs)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer reader.Close()

	// copy the relevant headers. If you want to preserve the downloaded file name, extract it with go's url parser.
	if attachment {
		// this causes the browser to download the content
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(pathAbs))
	}
	if strings.HasSuffix(pathAbs, ".json") {
		w.Header().Set("Content-Type", "application/json")
	}
	w.Header().Set("Content-Length", fmt.Sprint(fi.Size())) // useful for download progress

	// stream the body to the client without fully loading it into memory
	io.Copy(w, reader)
}

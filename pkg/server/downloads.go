package server

import (
	"net/http"

	"github.com/xbapps/xbvr/pkg/common"
)

type DownloadHandler struct {
}

func (h DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveFileFromDir(w, r, common.DownloadDir, true)
}

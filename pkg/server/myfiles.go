package server

import (
	"net/http"

	"github.com/xbapps/xbvr/pkg/common"
)

type MyFilesHandler struct {
}

func (h MyFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveFileFromDir(w, r, common.MyFilesDir, false)
}

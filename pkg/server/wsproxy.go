package server

import (
	"net/http"
	"net/url"
	"strings"
)

// wsOriginAllowed reports whether a websocket upgrade request may proceed.
// Requests carrying an Origin header (i.e. initiated by a web page) must be
// same-origin with the XBVR UI: the Origin host must equal the request Host.
// Requests without an Origin header (non-browser clients) are allowed.
func wsOriginAllowed(req *http.Request) bool {
	origin := req.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, req.Host)
}

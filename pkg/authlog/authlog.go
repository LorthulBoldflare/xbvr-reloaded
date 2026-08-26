// Package authlog is a standalone debug-logging facility for authentication
// and player-protocol traffic. It is intentionally self-contained (standard
// library only) so it can be used from any layer without import cycles.
//
// The log is OPT-IN: entries are only written when the XBVR_AUTH_LOG
// environment variable names a destination file; otherwise they are
// discarded. Output is one JSON object per line (JSONL). Entries include
// request headers and payloads verbatim — the file may contain credentials
// (player passwords, session tokens, Basic auth headers) and must be
// treated as sensitive.
package authlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// maxPayloadBytes caps how much of a request body is logged.
const maxPayloadBytes = 8192

// entry is one JSONL record. Kind discriminates: "meta", "event", "request".
type entry struct {
	TS            string      `json:"ts"`
	Component     string      `json:"component"`
	Kind          string      `json:"kind"`
	Message       string      `json:"message,omitempty"`
	Method        string      `json:"method,omitempty"`
	Path          string      `json:"path,omitempty"`
	Remote        string      `json:"remote,omitempty"`
	Headers       http.Header `json:"headers,omitempty"`
	Body          string      `json:"body,omitempty"`
	BodyTruncated bool        `json:"bodyTruncated,omitempty"`
}

var (
	mu     sync.Mutex
	opened bool // whether the "log opened" meta entry has been written
)

// write marshals and appends one entry. When XBVR_AUTH_LOG is unset or
// empty, the entry is discarded. All failures are silently ignored: debug
// logging must never break request handling.
func write(e entry) {
	p := os.Getenv("XBVR_AUTH_LOG")
	if p == "" {
		return
	}
	e.TS = time.Now().Format(time.RFC3339)

	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	if !opened {
		opened = true
		meta, _ := json.Marshal(entry{
			TS:        e.TS,
			Component: "authlog",
			Kind:      "meta",
			Message:   "xbvr auth log opened — WARNING: entries may contain credentials (passwords, tokens, Basic auth headers)",
		})
		f.Write(append(meta, '\n'))
	}

	line, err := json.Marshal(e)
	if err != nil {
		return
	}
	f.Write(append(line, '\n'))
}

// Event logs a single-line event (e.g. "auth method=basic user=admin result=success").
func Event(component, format string, args ...interface{}) {
	write(entry{
		Component: component,
		Kind:      "event",
		Message:   fmt.Sprintf(format, args...),
	})
}

// Request logs an incoming request: method, path, remote address, headers,
// and (capped) body. The caller is responsible for reading and restoring
// the body before calling this. Response payloads are deliberately not
// logged: too verbose. Auth outcomes are still visible via Event entries
// (success/failed/denied).
func Request(component string, r *http.Request, body []byte) {
	e := entry{
		Component: component,
		Kind:      "request",
		Method:    r.Method,
		Path:      r.URL.RequestURI(),
		Remote:    r.RemoteAddr,
		Headers:   r.Header,
	}
	if len(body) > 0 {
		if len(body) > maxPayloadBytes {
			e.Body = string(body[:maxPayloadBytes])
			e.BodyTruncated = true
		} else {
			e.Body = string(body)
		}
	}
	write(e)
}

// Package authlog is a standalone debug-logging facility for authentication
// and player-protocol traffic. It is intentionally self-contained (standard
// library only) so it can be used from any layer without import cycles.
//
// The log is OPT-IN: entries are only written when the XBVR_AUTH_LOG
// environment variable names a destination file; otherwise they are
// discarded. Output is one JSON object per line (JSONL), exactly one line
// per request. Entries include request headers and payloads verbatim — the
// file may contain credentials (player passwords, session tokens, Basic
// auth headers) and must be treated as sensitive.
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

// Entry is one JSONL record, built up over the lifetime of a single request
// and written exactly once via Done. Response payloads are deliberately not
// logged: too verbose.
type Entry struct {
	TS        string      `json:"ts"`
	Component string      `json:"component"`
	Method    string      `json:"method"`
	Path      string      `json:"path"`
	Remote    string      `json:"remote,omitempty"`
	Headers   http.Header `json:"headers,omitempty"`
	Body      string      `json:"body,omitempty"`
	// BodyTruncated marks that Body was capped at maxPayloadBytes.
	BodyTruncated bool `json:"bodyTruncated,omitempty"`

	// AuthMethod: how the request was (or could be) authenticated:
	// "cookie" (player session), "basic-ui", "basic-player", "form"
	// (login page), "protocol-body" (DeoVR/HereSphere native login),
	// "none".
	AuthMethod string `json:"authMethod,omitempty"`
	AuthUser   string `json:"authUser,omitempty"`
	// AuthResult: "accepted" (pre-existing session), "success" (fresh
	// credentials verified), "failed" (wrong credentials), "denied" (no or
	// unusable credentials, request rejected).
	AuthResult string `json:"authResult,omitempty"`

	// PresentedBasicUser is set when the request carries an Authorization:
	// Basic header (regardless of whether the credentials were accepted).
	PresentedBasicUser string `json:"presentedBasicUser,omitempty"`
	// PresentedPlayerCookie is "valid" or "invalid" when the request
	// carries the player session cookie, empty when absent.
	PresentedPlayerCookie string `json:"presentedPlayerCookie,omitempty"`

	// CookieMinted marks that the response sets the player session cookie.
	CookieMinted bool `json:"cookieMinted,omitempty"`
	// PlayerClient marks that the request matched the known-VR-player
	// detector (User-Agent / X-Requested-With).
	PlayerClient bool `json:"playerClient,omitempty"`
	// RedirectTo is set when the request is answered with a redirect.
	RedirectTo string   `json:"redirectTo,omitempty"`
	Notes      []string `json:"notes,omitempty"`
}

// Note appends a free-form remark to the entry (kept for exceptional
// circumstances; structured fields are preferred).
func (e *Entry) Note(format string, args ...interface{}) {
	e.Notes = append(e.Notes, fmt.Sprintf(format, args...))
}

// Done writes the entry as a single JSONL line. Call it exactly once per
// request (typically deferred at the start of the handler/filter).
func (e *Entry) Done() {
	e.TS = time.Now().Format(time.RFC3339)
	write(e)
}

var (
	mu     sync.Mutex
	opened bool // whether the "log opened" meta entry has been written
)

// Start begins an entry for one request. The caller is responsible for
// reading and restoring the body before calling this.
func Start(component string, r *http.Request, body []byte) *Entry {
	e := &Entry{
		Component: component,
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
	return e
}

// write marshals and appends one entry. When XBVR_AUTH_LOG is unset or
// empty, the entry is discarded. All failures are silently ignored: debug
// logging must never break request handling.
func write(e *Entry) {
	p := os.Getenv("XBVR_AUTH_LOG")
	if p == "" {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	if !opened {
		opened = true
		meta, _ := json.Marshal(&Entry{
			TS:        e.TS,
			Component: "authlog",
			Notes:     []string{"xbvr auth log opened — WARNING: entries may contain credentials (passwords, tokens, Basic auth headers)"},
		})
		f.Write(append(meta, '\n'))
	}

	line, err := json.Marshal(e)
	if err != nil {
		return
	}
	f.Write(append(line, '\n'))
}

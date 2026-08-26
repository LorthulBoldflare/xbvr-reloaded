package authlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupLogFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "auth.log")
	t.Setenv("XBVR_AUTH_LOG", p)
	return p
}

func readEntries(t *testing.T, p string) []Entry {
	t.Helper()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	var entries []Entry
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("line is not valid JSON: %q: %v", line, err)
		}
		entries = append(entries, e)
	}
	return entries
}

func TestRequestEntry(t *testing.T) {
	p := setupLogFile(t)

	req := httptest.NewRequest(http.MethodPost, "http://xbvr.local/deovr/?x=1", strings.NewReader("login=player"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	e := Start("deovr", req, []byte("login=player"))
	e.AuthMethod = "protocol-body"
	e.AuthUser = "player"
	e.AuthResult = "success"
	e.CookieMinted = true
	e.Done()

	entries := readEntries(t, p)
	// First entry of a fresh file is the meta warning, then the request.
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (meta + request), got %d: %+v", len(entries), entries)
	}
	if entries[0].Component != "authlog" || len(entries[0].Notes) == 0 ||
		!strings.Contains(entries[0].Notes[0], "may contain credentials") {
		t.Fatalf("first entry should be the sensitivity meta entry: %+v", entries[0])
	}
	got := entries[1]
	if got.Component != "deovr" || got.Method != "POST" || got.Path != "/deovr/?x=1" {
		t.Fatalf("unexpected request fields: %+v", got)
	}
	if got.Headers.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Fatalf("headers not preserved: %+v", got.Headers)
	}
	if got.Body != "login=player" || got.BodyTruncated {
		t.Fatalf("unexpected body fields: %+v", got)
	}
	if got.AuthMethod != "protocol-body" || got.AuthUser != "player" || got.AuthResult != "success" || !got.CookieMinted {
		t.Fatalf("unexpected auth fields: %+v", got)
	}
	if got.TS == "" {
		t.Fatal("entry missing timestamp")
	}
}

func TestOneRequestOneLine(t *testing.T) {
	p := setupLogFile(t)

	req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil)
	e := Start("api", req, nil)
	e.AuthMethod = "cookie"
	e.AuthResult = "accepted"
	e.Note("anything extra goes to notes")
	e.Done()

	// Exactly one non-meta line for the request, regardless of whether the
	// process-global meta preamble has already been written by another test.
	var requests []Entry
	for _, e := range readEntries(t, p) {
		if e.Component != "authlog" {
			requests = append(requests, e)
		}
	}
	if len(requests) != 1 {
		t.Fatalf("expected exactly 1 request line, got %d", len(requests))
	}
	got := requests[0]
	if len(got.Notes) != 1 || got.Notes[0] != "anything extra goes to notes" {
		t.Fatalf("notes not recorded: %+v", got.Notes)
	}
}

func TestRequestBodyTruncation(t *testing.T) {
	p := setupLogFile(t)

	req := httptest.NewRequest(http.MethodPost, "http://xbvr.local/x", nil)
	Start("test", req, []byte(strings.Repeat("x", maxPayloadBytes+100))).Done()

	entries := readEntries(t, p)
	got := entries[len(entries)-1]
	if len(got.Body) != maxPayloadBytes || !got.BodyTruncated {
		t.Fatalf("expected %d-byte truncated body, got len=%d truncated=%v", maxPayloadBytes, len(got.Body), got.BodyTruncated)
	}
}

func TestOptInNoPathDiscards(t *testing.T) {
	// No XBVR_AUTH_LOG and a HOME we control: nothing may be written anywhere.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XBVR_AUTH_LOG", "")

	Start("test", httptest.NewRequest(http.MethodGet, "http://xbvr.local/", nil), nil).Done()

	matches, err := filepath.Glob(filepath.Join(home, "*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("expected no files under $HOME, got %v (err=%v)", matches, err)
	}
}

func TestUnwritablePathNoPanic(t *testing.T) {
	t.Setenv("XBVR_AUTH_LOG", filepath.Join(t.TempDir(), "nonexistent-dir", "auth.log"))
	// Must not panic or error out even when the file cannot be opened.
	Start("test", httptest.NewRequest(http.MethodGet, "http://xbvr.local/", nil), nil).Done()
}

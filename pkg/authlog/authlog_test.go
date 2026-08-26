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

func readEntries(t *testing.T, p string) []entry {
	t.Helper()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	var entries []entry
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var e entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("line is not valid JSON: %q: %v", line, err)
		}
		entries = append(entries, e)
	}
	return entries
}

func TestEvent(t *testing.T) {
	p := setupLogFile(t)
	Event("test", "auth basic user=%q result=%s", "admin", "success")

	entries := readEntries(t, p)
	// First entry of a fresh file is the meta warning, then the event.
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (meta + event), got %d: %+v", len(entries), entries)
	}
	if entries[0].Kind != "meta" || !strings.Contains(entries[0].Message, "may contain credentials") {
		t.Fatalf("first entry should be the sensitivity meta entry: %+v", entries[0])
	}
	e := entries[1]
	if e.Kind != "event" || e.Component != "test" || e.Message != `auth basic user="admin" result=success` {
		t.Fatalf("unexpected event entry: %+v", e)
	}
	if e.TS == "" {
		t.Fatal("entry missing timestamp")
	}
}

func TestRequest(t *testing.T) {
	p := setupLogFile(t)

	req := httptest.NewRequest(http.MethodPost, "http://xbvr.local/deovr/?x=1", strings.NewReader("login=player"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	Request("deovr", req, []byte("login=player"))

	entries := readEntries(t, p)
	var got *entry
	for i := range entries {
		if entries[i].Kind == "request" {
			got = &entries[i]
		}
	}
	if got == nil {
		t.Fatalf("no request entry in %+v", entries)
	}
	if got.Component != "deovr" || got.Method != "POST" || got.Path != "/deovr/?x=1" {
		t.Fatalf("unexpected request entry: %+v", got)
	}
	if got.Headers.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Fatalf("headers not preserved: %+v", got.Headers)
	}
	if got.Body != "login=player" || got.BodyTruncated {
		t.Fatalf("unexpected body fields: %+v", got)
	}
}

func TestRequestBodyTruncation(t *testing.T) {
	p := setupLogFile(t)

	req := httptest.NewRequest(http.MethodPost, "http://xbvr.local/x", nil)
	Request("test", req, []byte(strings.Repeat("x", maxPayloadBytes+100)))

	entries := readEntries(t, p)
	var got *entry
	for i := range entries {
		if entries[i].Kind == "request" {
			got = &entries[i]
		}
	}
	if got == nil {
		t.Fatal("no request entry")
	}
	if len(got.Body) != maxPayloadBytes || !got.BodyTruncated {
		t.Fatalf("expected %d-byte truncated body, got len=%d truncated=%v", maxPayloadBytes, len(got.Body), got.BodyTruncated)
	}
}

func TestOptInNoPathDiscards(t *testing.T) {
	// No XBVR_AUTH_LOG and a HOME we control: nothing may be written anywhere.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XBVR_AUTH_LOG", "")

	Event("test", "discarded")

	matches, err := filepath.Glob(filepath.Join(home, "*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("expected no files under $HOME, got %v (err=%v)", matches, err)
	}
}

func TestUnwritablePathNoPanic(t *testing.T) {
	t.Setenv("XBVR_AUTH_LOG", filepath.Join(t.TempDir(), "nonexistent-dir", "auth.log"))
	// Must not panic or error out even when the file cannot be opened.
	Event("test", "unwritable")
}

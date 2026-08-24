package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServeFileFromDir(t *testing.T) {
	baseDir := t.TempDir()

	// legit nested file
	nested := filepath.Join(baseDir, "sub", "scene.json")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	// file outside the base dir
	outside := filepath.Join(filepath.Dir(baseDir), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outside)

	serve := func(target string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		rec := httptest.NewRecorder()
		serveFileFromDir(rec, req, baseDir, false)
		return rec
	}

	t.Run("serves nested file", func(t *testing.T) {
		rec := serve("/sub/scene.json")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected application/json, got %q", ct)
		}
	})

	t.Run("rejects traversal", func(t *testing.T) {
		for _, target := range []string{
			"/../secret.txt",
			"/..%2F..%2Fsecret.txt",
			"/sub/../../secret.txt",
			"/%2e%2e/secret.txt",
		} {
			rec := serve(target)
			if rec.Code == http.StatusOK {
				t.Fatalf("traversal %q was served (200)", target)
			}
			if rec.Body.String() == "secret" {
				t.Fatalf("traversal %q leaked file contents", target)
			}
		}
	})

	t.Run("missing file is 404", func(t *testing.T) {
		rec := serve("/nope.txt")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("directory is 404", func(t *testing.T) {
		rec := serve("/sub")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("no fd leak under repeated requests", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			rec := serve("/sub/scene.json")
			if rec.Code != http.StatusOK {
				t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
			}
		}
	})
}

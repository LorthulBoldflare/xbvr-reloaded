package tasks

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func makeZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractZipSafe(t *testing.T) {
	t.Run("extracts regular entries", func(t *testing.T) {
		zipPath := makeZip(t, map[string]string{"ffmpeg": "#!/bin/fake\n"})
		dest := t.TempDir()
		if err := extractZipSafe(zipPath, dest); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(dest, "ffmpeg"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "#!/bin/fake\n" {
			t.Fatalf("unexpected content: %q", data)
		}
	})

	t.Run("rejects traversal entries", func(t *testing.T) {
		zipPath := makeZip(t, map[string]string{"../evil.sh": "pwned"})
		dest := t.TempDir()
		if err := extractZipSafe(zipPath, dest); err == nil {
			t.Fatal("expected traversal entry to be rejected")
		}
		if _, err := os.Stat(filepath.Join(dest, "..", "evil.sh")); err == nil {
			t.Fatal("traversal entry was written outside destination")
		}
	})
}

func TestVerifySHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bin")
	content := []byte("binary content")
	if err := os.WriteFile(path, content, 0o755); err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(content)
	if err := verifySHA256(path, hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("expected valid hash to pass, got %v", err)
	}

	if err := verifySHA256(path, "0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("expected hash mismatch to be rejected")
	}
}

func TestFfbinariesHashPinsCoverSupportedPlatforms(t *testing.T) {
	platforms := []string{"windows-32", "windows-64", "osx-64", "linux-32", "linux-64", "linux-armhf", "linux-arm64"}
	tools := []string{"ffmpeg", "ffprobe"}
	for _, p := range platforms {
		for _, tool := range tools {
			key := p + "/" + tool
			if _, ok := ffbinariesSHA256[key]; !ok {
				t.Fatalf("missing pinned SHA-256 for %s", key)
			}
		}
	}
}

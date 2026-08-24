package migrations

import (
	"encoding/json"
	"testing"
)

// TestReleaseMigrationsEmbedded ensures the release migration data files are
// bundled into the binary via go:embed and contain parseable mappings, so
// migrations relying on them work without xbvr_data on disk.
func TestReleaseMigrationsEmbedded(t *testing.T) {
	entries, err := releaseMigrations.ReadDir("release")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no release migration files embedded")
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := readReleaseMigrationFile(e.Name())
		if err != nil {
			t.Errorf("cannot read embedded migration file %s: %v", e.Name(), err)
			continue
		}
		var data struct {
			Mappings []interface{} `json:"mappings"`
		}
		if err := json.Unmarshal(b, &data); err != nil {
			t.Errorf("embedded migration file %s is not valid JSON: %v", e.Name(), err)
			continue
		}
		if len(data.Mappings) == 0 {
			t.Errorf("embedded migration file %s has no mappings", e.Name())
		}
	}
}

// TestReadReleaseMigrationFileMissing ensures unknown files report an error
// instead of silently returning empty data.
func TestReadReleaseMigrationFileMissing(t *testing.T) {
	if _, err := readReleaseMigrationFile("does-not-exist.json"); err == nil {
		t.Error("expected error for missing embedded migration file")
	}
}

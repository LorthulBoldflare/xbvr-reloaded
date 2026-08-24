package migrations

import (
	"embed"
	"path"
)

// releaseMigrations bundles the release migration data files (scene ID
// remapping lists, ...) into the binary, so a standalone executable can run
// all migrations without xbvr_data being present next to it or in the app
// directory.
//
//go:embed release
var releaseMigrations embed.FS

// readReleaseMigrationFile returns the contents of a bundled release
// migration data file.
func readReleaseMigrationFile(filename string) ([]byte, error) {
	return releaseMigrations.ReadFile(path.Join("release", filename))
}

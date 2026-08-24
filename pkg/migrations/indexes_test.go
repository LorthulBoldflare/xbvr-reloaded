package migrations

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// indexExistsMust is a test helper wrapping indexExists.
func indexExistsMust(t *testing.T, db *gorm.DB, table, name string) bool {
	t.Helper()
	exists, err := indexExists(db, table, name)
	if err != nil {
		t.Fatal(err)
	}
	return exists
}

func TestAddPerformanceIndexes(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// minimal tables
	for _, stmt := range []string{
		"CREATE TABLE histories (id integer primary key, scene_id integer)",
		"CREATE TABLE scenes (id integer primary key, scene_id text, is_available boolean, is_accessible boolean, is_hidden boolean, release_date datetime)",
		"CREATE TABLE files (id integer primary key, volume_id integer, type text)",
		"CREATE TABLE external_references (id integer primary key, external_source text)",
		"CREATE TABLE actions (id integer primary key, scene_id text)",
		"CREATE TABLE action_actors (id integer primary key, actor_id integer)",
		"CREATE INDEX idx_scenes_scene_id ON scenes (scene_id)",
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := addPerformanceIndexes(db); err != nil {
		t.Fatal(err)
	}

	// all indexes created
	for _, i := range performanceIndexes {
		if !indexExistsMust(t, db, i.table, i.name) {
			t.Errorf("index %s on %s missing", i.name, i.table)
		}
	}

	// single scene_id index is kept (not dropped as "duplicate")
	if !indexExistsMust(t, db, "scenes", "idx_scenes_scene_id") {
		t.Error("idx_scenes_scene_id should be kept when it is the only scene_id index")
	}

	// idempotent: running again must not error
	if err := addPerformanceIndexes(db); err != nil {
		t.Fatalf("second run not idempotent: %v", err)
	}
}

func TestAddPerformanceIndexesDropsDuplicateSceneIDIndex(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, stmt := range []string{
		"CREATE TABLE scenes (id integer primary key, scene_id text, is_available boolean, is_accessible boolean, is_hidden boolean, release_date datetime)",
		"CREATE TABLE histories (id integer primary key, scene_id integer)",
		"CREATE TABLE files (id integer primary key, volume_id integer, type text)",
		"CREATE TABLE external_references (id integer primary key, external_source text)",
		"CREATE TABLE actions (id integer primary key, scene_id text)",
		"CREATE TABLE action_actors (id integer primary key, actor_id integer)",
		"CREATE INDEX idx_scenes_scene_id ON scenes (scene_id)",
		"CREATE INDEX scenes_scene_id_second ON scenes (scene_id)",
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := addPerformanceIndexes(db); err != nil {
		t.Fatal(err)
	}

	if sceneIDIndexCount(db) != 1 {
		t.Fatalf("expected exactly 1 index on scenes.scene_id after dedup, got %d", sceneIDIndexCount(db))
	}
	if indexExistsMust(t, db, "scenes", "idx_scenes_scene_id") {
		t.Error("duplicate idx_scenes_scene_id should have been dropped")
	}
	if !indexExistsMust(t, db, "scenes", "scenes_scene_id_second") {
		t.Error("remaining scene_id index should be kept")
	}
}

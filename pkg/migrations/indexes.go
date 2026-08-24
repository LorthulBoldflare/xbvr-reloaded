package migrations

import (
	"github.com/jinzhu/gorm"
)

// performanceIndexes lists indexes added by migration 0088. These columns are
// filtered/joined on hot query paths (scene lists, status refreshes, file
// matching, history updates) but had no index.
var performanceIndexes = []struct {
	table  string
	name   string
	column string
}{
	{"histories", "idx_histories_scene_id", "scene_id"},
	{"scenes", "idx_scenes_is_available", "is_available"},
	{"scenes", "idx_scenes_is_accessible", "is_accessible"},
	{"scenes", "idx_scenes_is_hidden", "is_hidden"},
	{"scenes", "idx_scenes_release_date", "release_date"},
	{"files", "idx_files_volume_id", "volume_id"},
	{"files", "idx_files_type", "type"},
	{"external_references", "idx_external_references_external_source", "external_source"},
	{"actions", "idx_actions_scene_id", "scene_id"},
	{"action_actors", "idx_action_actors_actor_id", "actor_id"},
}

// indexExists reports whether an index with the given name exists on the
// table, for both sqlite and mysql. Errors are propagated — misdetecting an
// existing index as missing would fail the migration with a duplicate-name
// error and (via tlog.Fatalf in Migrate) block startup.
func indexExists(tx *gorm.DB, table, name string) (bool, error) {
	if tx.Dialect().GetName() == "mysql" {
		var cnt int
		if err := tx.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?", table, name).Row().Scan(&cnt); err != nil {
			return false, err
		}
		return cnt > 0, nil
	}
	rows, err := tx.Raw("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND name = ?", table, name).Rows()
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

// addIndexIfNotExists creates the index in a dialect-appropriate,
// idempotent way.
func addIndexIfNotExists(tx *gorm.DB, table, name, column string) error {
	exists, err := indexExists(tx, table, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return tx.Table(table).AddIndex(name, column).Error
}

// sceneIDIndexCount returns how many indexes on the scenes table have
// scene_id as their first column.
func sceneIDIndexCount(tx *gorm.DB) int {
	if tx.Dialect().GetName() == "mysql" {
		var cnt int
		tx.Raw("SELECT COUNT(DISTINCT index_name) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'scenes' AND column_name = 'scene_id' AND seq_in_index = 1").Row().Scan(&cnt)
		return cnt
	}
	count := 0
	rows, err := tx.Raw("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = 'scenes'").Rows()
	if err != nil {
		return 0
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	for _, n := range names {
		var col string
		// PRAGMA index_info returns columns in index order; first row's name
		// is the leading column
		r := tx.Raw("PRAGMA index_info(" + n + ")").Row()
		var seqno, cid int
		if err := r.Scan(&seqno, &cid, &col); err == nil && col == "scene_id" {
			count++
		}
	}
	return count
}

// addPerformanceIndexes is migration 0088. It is idempotent: indexes are only
// created when missing, so the migration body is safe to re-run.
func addPerformanceIndexes(tx *gorm.DB) error {
	for _, i := range performanceIndexes {
		if err := addIndexIfNotExists(tx, i.table, i.name, i.column); err != nil {
			return err
		}
	}

	// scenes.scene_id was indexed twice (gorm:"index" tag and the manual
	// CREATE INDEX in migration 0073); drop the duplicate if two indexes
	// cover the column, otherwise keep whichever exists
	if sceneIDIndexCount(tx) > 1 {
		// only drop when the index actually exists — the manual CREATE INDEX
		// error was ignored in 0073, so it may never have been created
		exists, err := indexExists(tx, "scenes", "idx_scenes_scene_id")
		if err != nil {
			return err
		}
		if exists {
			var drop string
			if tx.Dialect().GetName() == "mysql" {
				drop = "ALTER TABLE scenes DROP INDEX idx_scenes_scene_id"
			} else {
				drop = "DROP INDEX IF EXISTS idx_scenes_scene_id"
			}
			if err := tx.Exec(drop).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

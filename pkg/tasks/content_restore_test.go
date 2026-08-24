package tasks

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/xbapps/xbvr/pkg/models"
)

func setupRestoreActionsDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.AutoMigrate(&models.Action{}, &models.History{}, &models.Scene{}).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

// Regression test: RestoreActions with overwrite used to delete from the
// History table (wrong table, and a string scene id compared against the
// numeric history FK) instead of deleting the old Action rows.
func TestRestoreActionsOverwritePreservesHistory(t *testing.T) {
	db := setupRestoreActionsDB(t)

	scene := models.Scene{ID: 1, SceneID: "scene-1"}
	if err := db.Create(&scene).Error; err != nil {
		t.Fatal(err)
	}

	// existing watch history for the scene — must survive the restore
	history := models.History{SceneID: scene.ID}
	if err := db.Create(&history).Error; err != nil {
		t.Fatal(err)
	}

	// stale action rows that overwrite=true should replace
	oldAction := models.Action{SceneID: "scene-1", ActionType: "edit", ChangedColumn: "title", NewValue: "old"}
	if err := db.Create(&oldAction).Error; err != nil {
		t.Fatal(err)
	}

	bundle := []BackupSceneAction{
		{
			SceneID: "scene-1",
			Actions: []models.Action{
				{SceneID: "scene-1", ActionType: "edit", ChangedColumn: "title", NewValue: "new"},
			},
		},
	}
	RestoreActions(bundle, true, nil, true, db)

	// history preserved
	var historyCount int
	if err := db.Model(&models.History{}).Where("scene_id = ?", scene.ID).Count(&historyCount).Error; err != nil {
		t.Fatal(err)
	}
	if historyCount != 1 {
		t.Fatalf("expected 1 history row after restore, got %d", historyCount)
	}

	// old action replaced by the restored one
	var actions []models.Action
	if err := db.Where("scene_id = ?", "scene-1").Find(&actions).Error; err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action row after restore, got %d", len(actions))
	}
	if actions[0].NewValue != "new" {
		t.Fatalf("expected restored action NewValue=%q, got %q", "new", actions[0].NewValue)
	}
	if actions[0].ID == oldAction.ID {
		t.Fatal("old action row was not deleted")
	}
}

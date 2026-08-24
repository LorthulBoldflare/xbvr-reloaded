package tasks

import (
	"strings"
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

// Regression test for #2247: a malformed bundle (e.g. an empty string for a
// time.Time field) must be rejected by validation instead of being restored
// partially while reporting success.
func TestValidateBundle(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		wantErr    bool
		errContain string
	}{
		{
			name:    "valid 2.1 bundle",
			data:    `{"bundleVersion":"2.1","scenes":[{"scene_id":"example-1","title":"Example","cast":[{"name":"Example Actor","birth_date":"0001-01-01T00:00:00Z"}]}]}`,
			wantErr: false,
		},
		{
			name:    "empty birth_date string",
			data:    `{"bundleVersion":"2.1","scenes":[{"scene_id":"example-1","title":"Example","cast":[{"name":"Example Actor","birth_date":""}]}]}`,
			wantErr: true,
		},
		{
			name:       "wrong bundle version",
			data:       `{"bundleVersion":"2.0","scenes":[]}`,
			wantErr:    true,
			errContain: "version",
		},
		{
			name:    "valid v1 bundle",
			data:    `{"bundleVersion":"1","scenes":[]}`,
			wantErr: false,
		},
		{
			name:    "not JSON",
			data:    `this is not a bundle`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateBundle(tc.data)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tc.errContain != "" && (err == nil || !strings.Contains(err.Error(), tc.errContain)) {
				t.Fatalf("expected error containing %q, got %v", tc.errContain, err)
			}
		})
	}
}

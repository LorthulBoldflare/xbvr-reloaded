package models

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/markphelps/optional"
)

func TestOldestCreationTime(t *testing.T) {
	epoch := time.Unix(0, 0)
	oldestTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	newerTime := time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)

	files := []File{
		{Type: "video", CreatedTime: epoch, Volume: Volume{Type: "local"}},
		{Type: "video", CreatedTime: newerTime, Volume: Volume{Type: "local"}},
		{Type: "script", CreatedTime: oldestTime, Volume: Volume{Type: "local"}},
		{Type: "hsp", CreatedTime: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), Volume: Volume{Type: "local"}},
	}

	path := t.TempDir()
	filename := "file"
	if err := os.WriteFile(filepath.Join(path, filename), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	for i := range files {
		files[i].Path = path
		files[i].Filename = filename
	}

	if got := oldestCreationTime(files); !got.Equal(oldestTime) {
		t.Errorf("oldestCreationTime() = %v, want %v", got, oldestTime)
	}
}

func TestQueryScenesSortsByDuration(t *testing.T) {	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&Scene{}, &KV{}).Error; err != nil {
		t.Fatal(err)
	}

	scenes := []Scene{
		{ID: 1, SceneID: "scene-1", Duration: 30},
		{ID: 2, SceneID: "scene-2", Duration: 0},
		{ID: 3, SceneID: "scene-3", Duration: 10},
		{ID: 4, SceneID: "scene-4", Duration: 30},
		{ID: 5, SceneID: "scene-5", Duration: -1},
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name string
		sort string
		want []uint
	}{
		{
			name: "shortest first",
			sort: "duration_asc",
			want: []uint{3, 1, 4, 2, 5},
		},
		{
			name: "longest first",
			sort: "duration_desc",
			want: []uint{1, 4, 3, 2, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequestSceneList{Sort: optional.NewString(tt.sort)}
			_, tx := queryScenes(db, r)

			var got []uint
			if err := tx.Pluck("scenes.id", &got).Error; err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got IDs %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("got IDs %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}

// Regression test: re-scraping an existing scene whose scraped data contains
// timestamps must import them as cuepoints, replacing previous simple
// (track IS NULL) cuepoints while preserving HereSphere (track IS NOT NULL)
// cuepoints.
func TestSceneCreateUpdateFromExternalImportsCuepointsOnRescrape(t *testing.T) {
	setup := func(t *testing.T) (*gorm.DB, Scene) {
		db, err := gorm.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { db.Close() })

		if err := db.AutoMigrate(&Scene{}, &SceneCuepoint{}, &Tag{}, &Actor{}, &Site{}, &ExternalReference{}, &ExternalReferenceLink{}, &KV{}).Error; err != nil {
			t.Fatal(err)
		}

		trackOne := uint(1)
		existing := Scene{SceneID: "slr-1", Title: "T"}
		if err := db.Create(&existing).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&SceneCuepoint{SceneID: existing.ID, Name: "Old", TimeStart: 1}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&SceneCuepoint{SceneID: existing.ID, Name: "HSP", TimeStart: 2, Track: &trackOne}).Error; err != nil {
			t.Fatal(err)
		}
		return db, existing
	}

	getCuepoints := func(t *testing.T, db *gorm.DB, sceneID uint) []SceneCuepoint {
		var cuepoints []SceneCuepoint
		if err := db.Where("scene_id = ?", sceneID).Find(&cuepoints).Error; err != nil {
			t.Fatal(err)
		}
		return cuepoints
	}

	t.Run("timestamps replace simple cuepoints and preserve HSP cuepoints", func(t *testing.T) {
		db, existing := setup(t)

		// Title must match so SceneCreateUpdateFromExternal does not call
		// GetScriptFiles(), which uses the global DB handle.
		ext := ScrapedScene{
			SceneID:    "slr-1",
			Title:      "T",
			Timestamps: `[{"Intro":10},{"Cowgirl":120.5},{"Finale":"500:560"}]`,
		}
		if err := SceneCreateUpdateFromExternal(db, ext); err != nil {
			t.Fatal(err)
		}

		cuepoints := getCuepoints(t, db, existing.ID)
		if len(cuepoints) != 4 {
			t.Fatalf("expected 4 cuepoints after re-scrape, got %d: %+v", len(cuepoints), cuepoints)
		}

		byName := map[string]SceneCuepoint{}
		for _, cp := range cuepoints {
			byName[cp.Name] = cp
		}

		if _, found := byName["Old"]; found {
			t.Error("simple cuepoint 'Old' should have been replaced by scraped timestamps")
		}
		hsp, found := byName["HSP"]
		if !found {
			t.Fatal("HereSphere cuepoint 'HSP' should have been preserved")
		}
		if hsp.Track == nil || *hsp.Track != 1 {
			t.Errorf("HSP cuepoint track = %v, want 1", hsp.Track)
		}

		if cp := byName["Intro"]; cp.TimeStart != 10 || cp.TimeEnd != 0 || cp.Track != nil {
			t.Errorf("Intro cuepoint = %+v, want TimeStart=10 TimeEnd=0 Track=nil", cp)
		}
		if cp := byName["Cowgirl"]; cp.TimeStart != 120.5 || cp.Track != nil {
			t.Errorf("Cowgirl cuepoint = %+v, want TimeStart=120.5 Track=nil", cp)
		}
		if cp := byName["Finale"]; cp.TimeStart != 500 || cp.TimeEnd != 560 || cp.Track != nil {
			t.Errorf("Finale cuepoint = %+v, want TimeStart=500 TimeEnd=560 Track=nil", cp)
		}
	})

	t.Run("empty timestamps leave existing cuepoints untouched", func(t *testing.T) {
		db, existing := setup(t)

		ext := ScrapedScene{
			SceneID: "slr-1",
			Title:   "T",
		}
		if err := SceneCreateUpdateFromExternal(db, ext); err != nil {
			t.Fatal(err)
		}

		cuepoints := getCuepoints(t, db, existing.ID)
		if len(cuepoints) != 2 {
			t.Fatalf("expected 2 cuepoints to remain, got %d: %+v", len(cuepoints), cuepoints)
		}
	})
}

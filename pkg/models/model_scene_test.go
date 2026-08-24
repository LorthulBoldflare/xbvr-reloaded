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

func TestQueryScenesSortsByDuration(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
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

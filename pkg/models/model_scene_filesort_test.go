package models

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/markphelps/optional"
)

// Scenes must be orderable by the creation time of their newest video file;
// scenes without video files sort last in both directions.
func TestQueryScenesSortsByFileAdded(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&Scene{}, &File{}, &KV{}).Error; err != nil {
		t.Fatal(err)
	}

	scenes := []Scene{
		{ID: 1, SceneID: "scene-1"}, // newer file
		{ID: 2, SceneID: "scene-2"}, // older file
		{ID: 3, SceneID: "scene-3"}, // no files
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	files := []File{
		{SceneID: 1, Type: "video", CreatedTime: time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)},
		{SceneID: 2, Type: "video", CreatedTime: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)},
		// A newer non-video file must not affect the ordering.
		{SceneID: 2, Type: "script", CreatedTime: time.Date(2024, time.January, 5, 0, 0, 0, 0, time.UTC)},
	}
	for i := range files {
		if err := db.Create(&files[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name string
		sort string
		want []uint
	}{
		{name: "newest file first", sort: "file_added_desc", want: []uint{1, 2, 3}},
		{name: "oldest file first", sort: "file_added_asc", want: []uint{2, 1, 3}},
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

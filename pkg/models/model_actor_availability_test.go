package models

import (
	"sort"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/markphelps/optional"
)

// The actor availability presets (driven by the isAvailable/isAccessible flags
// on RequestActorList) filter actors by the availability of their associated
// scenes: "available" is a live subquery on scene_cast+scenes (avail_count may
// be stale between scrapes), while downloaded/not-downloaded use the
// denormalized avail_count column, which excludes hidden scenes.
//
// Note: tests must not set Attributes (GetCountryList uses the global DB) or
// sort "random" (reads the global dbConn).
func TestQueryActorsAvailability(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&Actor{}, &Scene{}).Error; err != nil {
		t.Fatal(err)
	}

	actors := []Actor{
		{ID: 1, Name: "accessible", Count: 1, AvailCount: 1},   // available + accessible scene
		{ID: 2, Name: "inaccessible", Count: 1, AvailCount: 1}, // available but volume offline
		{ID: 3, Name: "missing", Count: 1, AvailCount: 0},      // scene not downloaded
		{ID: 4, Name: "scenefree", Count: 0, AvailCount: 0},    // no scenes at all
		{ID: 5, Name: "hiddenonly", Count: 1, AvailCount: 0},   // only a hidden available scene
		{ID: 6, Name: "stalecount", Count: 1, AvailCount: 0},   // playable scene, stale avail_count
	}
	for i := range actors {
		if err := db.Create(&actors[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	scenes := []Scene{
		{ID: 1, SceneID: "scene-1", IsAvailable: true, IsAccessible: true},
		{ID: 2, SceneID: "scene-2", IsAvailable: true, IsAccessible: false},
		{ID: 3, SceneID: "scene-3", IsAvailable: false, IsAccessible: false},
		{ID: 5, SceneID: "scene-5", IsAvailable: true, IsAccessible: true, IsHidden: true},
		{ID: 6, SceneID: "scene-6", IsAvailable: true, IsAccessible: true},
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	cast := [][2]uint{{1, 1}, {2, 2}, {3, 3}, {5, 5}, {6, 6}} // actor_id, scene_id
	for _, c := range cast {
		if err := db.Exec("insert into scene_cast (actor_id, scene_id) values (?, ?)", c[0], c[1]).Error; err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name         string
		isAvailable  optional.Bool
		isAccessible optional.Bool
		want         []uint
	}{
		{name: "available right now is live and ignores avail_count",
			isAvailable: optional.NewBool(true), isAccessible: optional.NewBool(true), want: []uint{1, 6}},
		{name: "downloaded uses avail_count",
			isAvailable: optional.NewBool(true), want: []uint{1, 2}},
		// A stale avail_count shows the actor in both "available" (live
		// subquery) and "not downloaded" (column) until the next recount.
		{name: "not downloaded uses avail_count",
			isAvailable: optional.NewBool(false), want: []uint{3, 4, 5, 6}},
		{name: "any applies no availability filter",
			want: []uint{1, 2, 3, 4, 5, 6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequestActorList{IsAvailable: tt.isAvailable, IsAccessible: tt.isAccessible}
			_, tx := queryActors(db, r)

			var got []uint
			if err := tx.Pluck("actors.id", &got).Error; err != nil {
				t.Fatal(err)
			}
			sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
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

// The per-option availability badge counts mirror the preset clauses and are
// computed under the remaining filters (preCountTx).
func TestQueryActorsAvailabilityCounts(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&Actor{}, &Scene{}).Error; err != nil {
		t.Fatal(err)
	}

	actors := []Actor{
		{ID: 1, Name: "accessible", Count: 1, AvailCount: 1},
		{ID: 2, Name: "inaccessible", Count: 1, AvailCount: 1},
		{ID: 3, Name: "missing", Count: 1, AvailCount: 0},
		{ID: 4, Name: "scenefree", Count: 0, AvailCount: 0},
		{ID: 5, Name: "hiddenonly", Count: 1, AvailCount: 0},
		{ID: 6, Name: "stalecount", Count: 1, AvailCount: 0},
	}
	for i := range actors {
		if err := db.Create(&actors[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	scenes := []Scene{
		{ID: 1, SceneID: "scene-1", IsAvailable: true, IsAccessible: true},
		{ID: 2, SceneID: "scene-2", IsAvailable: true, IsAccessible: false},
		{ID: 3, SceneID: "scene-3", IsAvailable: false, IsAccessible: false},
		{ID: 5, SceneID: "scene-5", IsAvailable: true, IsAccessible: true, IsHidden: true},
		{ID: 6, SceneID: "scene-6", IsAvailable: true, IsAccessible: true},
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	cast := [][2]uint{{1, 1}, {2, 2}, {3, 3}, {5, 5}, {6, 6}}
	for _, c := range cast {
		if err := db.Exec("insert into scene_cast (actor_id, scene_id) values (?, ?)", c[0], c[1]).Error; err != nil {
			t.Fatal(err)
		}
	}

	// Even with an availability filter active, counts reflect all options.
	r := RequestActorList{IsAvailable: optional.NewBool(true), IsAccessible: optional.NewBool(true)}
	preCountTx, _ := queryActors(db, r)

	var out ResponseActorList
	countActorAvailability(preCountTx, &out)

	if out.CountAny != 6 {
		t.Errorf("CountAny = %d, want 6", out.CountAny)
	}
	if out.CountAvailable != 2 {
		t.Errorf("CountAvailable = %d, want 2 (accessible + stalecount)", out.CountAvailable)
	}
	if out.CountDownloaded != 2 {
		t.Errorf("CountDownloaded = %d, want 2 (accessible + inaccessible)", out.CountDownloaded)
	}
	if out.CountNotDownloaded != 4 {
		t.Errorf("CountNotDownloaded = %d, want 4 (missing, scenefree, hiddenonly, stalecount)", out.CountNotDownloaded)
	}
	if out.CountHidden != 0 {
		t.Errorf("CountHidden = %d, want 0 (actors have no hidden flag)", out.CountHidden)
	}
}

// avail_count must exclude hidden scenes while count keeps them.
func TestCountActorTagsExcludesHiddenScenes(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&Actor{}, &Scene{}).Error; err != nil {
		t.Fatal(err)
	}

	actors := []Actor{
		{ID: 1, Name: "mixed"},      // one visible + one hidden available scene
		{ID: 2, Name: "hiddenonly"}, // only a hidden available scene
	}
	for i := range actors {
		if err := db.Create(&actors[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	scenes := []Scene{
		{ID: 1, SceneID: "scene-1", IsAvailable: true, IsAccessible: true},
		{ID: 2, SceneID: "scene-2", IsAvailable: true, IsAccessible: true, IsHidden: true},
		{ID: 3, SceneID: "scene-3", IsAvailable: true, IsAccessible: true, IsHidden: true},
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	cast := [][2]uint{{1, 1}, {1, 2}, {2, 3}}
	for _, c := range cast {
		if err := db.Exec("insert into scene_cast (actor_id, scene_id) values (?, ?)", c[0], c[1]).Error; err != nil {
			t.Fatal(err)
		}
	}

	countActorTags(db)

	var got []Actor
	if err := db.Order("id").Find(&got).Error; err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d actors, want 2", len(got))
	}
	if got[0].Count != 2 || got[0].AvailCount != 1 {
		t.Errorf("mixed: count=%d avail_count=%d, want 2/1", got[0].Count, got[0].AvailCount)
	}
	if got[1].Count != 1 || got[1].AvailCount != 0 {
		t.Errorf("hiddenonly: count=%d avail_count=%d, want 1/0", got[1].Count, got[1].AvailCount)
	}
}

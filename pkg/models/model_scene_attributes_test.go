package models

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/markphelps/optional"
)

// sqlLogger captures the SQL statements gorm generates.
type sqlLogger struct {
	queries *[]string
}

func (l sqlLogger) Print(v ...interface{}) {
	if len(v) > 3 && v[0] == "sql" {
		*l.queries = append(*l.queries, fmt.Sprintf("%v", v[3]))
	}
}

func setupAttributeFilterDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.AutoMigrate(&Scene{}, &File{}, &Volume{}, &KV{}).Error; err != nil {
		t.Fatal(err)
	}

	scenes := []Scene{
		{ID: 1, SceneID: "scene-1", StarRating: 4},
		{ID: 2, SceneID: "scene-2", StarRating: 2},
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	vol := Volume{ID: 1, Type: "local"}
	if err := db.Create(&vol).Error; err != nil {
		t.Fatal(err)
	}

	files := []File{
		// 3840 wide => (3840+500)/1000 = 4 (integer division) => "4K"
		{ID: 1, SceneID: 1, VolumeID: 1, Type: "video", VideoWidth: 3840, VideoCodecName: "h264", VideoAvgFrameRateVal: 60},
		{ID: 2, SceneID: 2, VolumeID: 1, Type: "video", VideoWidth: 1920, VideoCodecName: "hevc", VideoAvgFrameRateVal: 30},
	}
	for i := range files {
		if err := db.Create(&files[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	return db
}

func queryAttributeIDs(t *testing.T, db *gorm.DB, attrs ...string) ([]uint, []string) {
	t.Helper()

	var queries []string
	db.LogMode(true)
	db.SetLogger(sqlLogger{queries: &queries})
	defer db.LogMode(false)

	r := RequestSceneList{}
	for _, a := range attrs {
		r.Attributes = append(r.Attributes, optional.NewString(a))
	}

	_, tx := queryScenes(db, r)

	var ids []uint
	if err := tx.Pluck("scenes.id", &ids).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}
	return ids, queries
}

func idsEqual(got []uint, want ...uint) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func assertNoInjection(t *testing.T, queries []string, payload string) {
	t.Helper()
	for _, q := range queries {
		if strings.Contains(q, payload) {
			t.Fatalf("payload %q found in generated SQL: %s", payload, q)
		}
	}
}

func TestSceneAttributeFilters(t *testing.T) {
	db := setupAttributeFilterDB(t)

	t.Run("rating filter works", func(t *testing.T) {
		ids, _ := queryAttributeIDs(t, db, "Rating 4")
		if !idsEqual(ids, 1) {
			t.Fatalf("got IDs %v, want [1]", ids)
		}
	})

	t.Run("resolution filter works", func(t *testing.T) {
		ids, _ := queryAttributeIDs(t, db, "Resolution 4K")
		if !idsEqual(ids, 1) {
			t.Fatalf("got IDs %v, want [1]", ids)
		}
	})

	t.Run("frame rate filter works", func(t *testing.T) {
		ids, _ := queryAttributeIDs(t, db, "Frame Rate 60fps")
		if !idsEqual(ids, 1) {
			t.Fatalf("got IDs %v, want [1]", ids)
		}
	})

	t.Run("codec filter works", func(t *testing.T) {
		ids, _ := queryAttributeIDs(t, db, "Codec hevc")
		if !idsEqual(ids, 2) {
			t.Fatalf("got IDs %v, want [2]", ids)
		}
	})

	t.Run("malicious rating is inert", func(t *testing.T) {
		ids, queries := queryAttributeIDs(t, db, "Rating 0) OR 1=1--")
		assertNoInjection(t, queries, "OR 1=1")
		// filter dropped => both scenes returned, no error
		if !idsEqual(ids, 1, 2) {
			t.Fatalf("got IDs %v, want [1 2]", ids)
		}
	})

	t.Run("malicious resolution is inert", func(t *testing.T) {
		ids, queries := queryAttributeIDs(t, db, "Resolution 4K) OR 1=1--")
		assertNoInjection(t, queries, "OR 1=1")
		if !idsEqual(ids, 1, 2) {
			t.Fatalf("got IDs %v, want [1 2]", ids)
		}
	})

	t.Run("malicious frame rate is inert", func(t *testing.T) {
		ids, queries := queryAttributeIDs(t, db, "Frame Rate 60) OR 1=1--")
		assertNoInjection(t, queries, "OR 1=1")
		if !idsEqual(ids, 1, 2) {
			t.Fatalf("got IDs %v, want [1 2]", ids)
		}
	})

	t.Run("malicious codec is inert", func(t *testing.T) {
		ids, queries := queryAttributeIDs(t, db, "Codec h264') OR ('1'='1")
		assertNoInjection(t, queries, "'1'='1")
		if !idsEqual(ids, 1, 2) {
			t.Fatalf("got IDs %v, want [1 2]", ids)
		}
	})

	t.Run("malicious negated rating is inert", func(t *testing.T) {
		ids, queries := queryAttributeIDs(t, db, "!Rating 4) OR 1=1--")
		assertNoInjection(t, queries, "OR 1=1")
		if !idsEqual(ids, 1, 2) {
			t.Fatalf("got IDs %v, want [1 2]", ids)
		}
	})
}

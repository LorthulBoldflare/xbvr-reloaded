package models

import "testing"

// TestGetIfExistByURL verifies the exact-match scene URL lookup used by
// single-scene scraping: it tolerates a trailing-slash difference but must
// not match on LIKE wildcards (% or _) or path prefixes, which the previous
// `scene_url LIKE 'url%'` query allowed.
func TestGetIfExistByURL(t *testing.T) {
	db, err := GetCommonDB()
	if err != nil {
		t.Fatalf("get db: %v", err)
	}
	if err := db.AutoMigrate(&Scene{}, &Tag{}, &Actor{}, &File{},
		&Volume{}, &Action{}, &History{}, &SceneCuepoint{}).Error; err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	scene := Scene{SceneID: "urlmatch-scene-1", Title: "URL Match Test",
		SceneURL: "https://example.invalid/urlmatch/scene-1"}
	if err := db.Create(&scene).Error; err != nil {
		t.Fatal(err)
	}

	t.Run("exact match", func(t *testing.T) {
		var got Scene
		if err := got.GetIfExistByURL(scene.SceneURL); err != nil {
			t.Fatalf("expected match: %v", err)
		}
		if got.ID != scene.ID {
			t.Fatalf("matched scene %d, want %d", got.ID, scene.ID)
		}
	})

	t.Run("trailing slash tolerated", func(t *testing.T) {
		var got Scene
		if err := got.GetIfExistByURL(scene.SceneURL + "/"); err != nil {
			t.Fatalf("expected match with trailing slash: %v", err)
		}
		if got.ID != scene.ID {
			t.Fatalf("matched scene %d, want %d", got.ID, scene.ID)
		}
	})

	t.Run("path prefix does not match", func(t *testing.T) {
		var got Scene
		if err := got.GetIfExistByURL("https://example.invalid/urlmatch/scene"); err == nil {
			t.Fatalf("prefix URL must not match, got scene %d", got.ID)
		}
	})

	t.Run("LIKE wildcard does not match", func(t *testing.T) {
		var got Scene
		if err := got.GetIfExistByURL("https://example.invalid/urlmatch/%"); err == nil {
			t.Fatalf("wildcard URL must not match, got scene %d", got.ID)
		}
		var gotUnderscore Scene
		if err := gotUnderscore.GetIfExistByURL("https://example.invalid/urlmatch/scene-_"); err == nil {
			t.Fatalf("underscore wildcard must not match, got scene %d", gotUnderscore.ID)
		}
	})
}

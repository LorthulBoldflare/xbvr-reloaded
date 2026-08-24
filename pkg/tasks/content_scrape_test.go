package tasks

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/xbapps/xbvr/pkg/models"
)

// Scenes flagged needs_update must cause their site to be treated as having
// pending updates, so runScrapers can bypass limit scraping for that site and
// the scenes actually get re-scraped.
func TestSitesWithPendingUpdates(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.AutoMigrate(&models.Scene{}).Error; err != nil {
		t.Fatal(err)
	}

	scenes := []models.Scene{
		{SceneID: "scene-a1", ScraperId: "site-a", NeedsUpdate: true},
		{SceneID: "scene-a2", ScraperId: "site-a", NeedsUpdate: false},
		{SceneID: "scene-b1", ScraperId: "site-b", NeedsUpdate: false},
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	pending := sitesWithPendingUpdates(db)

	if len(pending) != 1 {
		t.Fatalf("expected 1 site with pending updates, got %d: %v", len(pending), pending)
	}
	if !pending["site-a"] {
		t.Error("expected site-a to have pending updates")
	}
	if pending["site-b"] {
		t.Error("site-b has no scenes flagged needs_update and must not be pending")
	}

	// clearing the flag clears the pending state
	if err := db.Model(&models.Scene{}).Where("scene_id = ?", "scene-a1").Update("needs_update", false).Error; err != nil {
		t.Fatal(err)
	}
	if pending := sitesWithPendingUpdates(db); len(pending) != 0 {
		t.Errorf("expected no pending sites after clearing needs_update, got %v", pending)
	}
}

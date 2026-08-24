package tasks

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/xbapps/xbvr/pkg/models"
)

func setupUpdateLinksDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.AutoMigrate(&models.ExternalReference{}, &models.ExternalReferenceLink{}).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func listLinks(t *testing.T, db *gorm.DB, extRefID uint) []models.ExternalReferenceLink {
	t.Helper()
	var links []models.ExternalReferenceLink
	if err := db.Where("external_reference_id = ?", extRefID).Find(&links).Error; err != nil {
		t.Fatal(err)
	}
	return links
}

func TestUpdateLinks(t *testing.T) {
	db := setupUpdateLinksDB(t)

	extref := models.ExternalReference{ExternalSource: "alternate scene test", ExternalId: "test-1"}
	if err := db.Create(&extref).Error; err != nil {
		t.Fatal(err)
	}

	newLink := models.ExternalReferenceLink{
		InternalTable:       "scenes",
		InternalDbId:        42,
		InternalNameId:      "scene-42",
		ExternalReferenceID: extref.ID,
		ExternalSource:      extref.ExternalSource,
		ExternalId:          extref.ExternalId,
		MatchType:           10000,
		UdfDatetime1:        time.Now(),
	}

	t.Run("creates link when none exist", func(t *testing.T) {
		UpdateLinks(db, extref.ID, newLink)
		links := listLinks(t, db, extref.ID)
		if len(links) != 1 || links[0].InternalDbId != 42 {
			t.Fatalf("expected exactly one link to scene 42, got %+v", links)
		}
	})

	t.Run("does not duplicate existing link", func(t *testing.T) {
		UpdateLinks(db, extref.ID, newLink)
		UpdateLinks(db, extref.ID, newLink)
		links := listLinks(t, db, extref.ID)
		if len(links) != 1 {
			t.Fatalf("expected link to remain deduplicated, got %+v", links)
		}
	})

	t.Run("replaces stale links", func(t *testing.T) {
		stale := models.ExternalReferenceLink{
			InternalTable:       "scenes",
			InternalDbId:        7,
			InternalNameId:      "scene-7",
			ExternalReferenceID: extref.ID,
			ExternalSource:      extref.ExternalSource,
			ExternalId:          extref.ExternalId,
		}
		if err := db.Create(&stale).Error; err != nil {
			t.Fatal(err)
		}
		otherRefLink := models.ExternalReferenceLink{
			InternalTable:       "scenes",
			InternalDbId:        7,
			InternalNameId:      "scene-7",
			ExternalReferenceID: extref.ID + 100, // belongs to another external reference
			ExternalSource:      extref.ExternalSource,
			ExternalId:          "other",
		}
		if err := db.Create(&otherRefLink).Error; err != nil {
			t.Fatal(err)
		}

		UpdateLinks(db, extref.ID, newLink)

		links := listLinks(t, db, extref.ID)
		if len(links) != 1 || links[0].InternalDbId != 42 {
			t.Fatalf("expected stale link replaced by link to 42, got %+v", links)
		}
		// link belonging to another external reference must be untouched
		var untouched models.ExternalReferenceLink
		if err := db.Where("external_reference_id = ?", otherRefLink.ExternalReferenceID).First(&untouched).Error; err != nil {
			t.Fatalf("link of another external reference was deleted: %v", err)
		}
	})
}

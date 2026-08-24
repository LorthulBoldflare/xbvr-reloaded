package api

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/xbapps/xbvr/pkg/models"
)

func TestQueryFilters(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.AutoMigrate(&models.Scene{}, &models.Tag{}, &models.Actor{}, &models.File{}, &models.Volume{}).Error; err != nil {
		t.Fatal(err)
	}

	tag1 := models.Tag{Name: "tag-one"}
	tag2 := models.Tag{Name: "tag-two"}
	db.Create(&tag1)
	db.Create(&tag2)
	actor1 := models.Actor{Name: "Actor One"}
	db.Create(&actor1)

	release := time.Date(2023, time.March, 15, 0, 0, 0, 0, time.UTC)
	scenes := []models.Scene{
		{ID: 1, SceneID: "s1", Site: "SiteA", IsAvailable: true, IsAccessible: true, ReleaseDate: release, Tags: []models.Tag{tag1}, Cast: []models.Actor{actor1}},
		{ID: 2, SceneID: "s2", Site: "SiteA", IsAvailable: true, IsAccessible: true, ReleaseDate: release, Tags: []models.Tag{tag1, tag2}},
		{ID: 3, SceneID: "s3", Site: "SiteB", IsAvailable: false, IsAccessible: true, ReleaseDate: release.AddDate(0, 1, 0)},
	}
	for i := range scenes {
		if err := db.Create(&scenes[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	boolPtr := func(b bool) *bool { return &b }

	t.Run("no filters", func(t *testing.T) {
		sites, tags, cast, months := queryFilters(db, nil, nil)
		if len(sites) != 2 {
			t.Errorf("sites = %v, want 2 entries", sites)
		}
		if len(tags) != 2 {
			t.Errorf("tags = %v, want 2 entries", tags)
		}
		if len(cast) != 1 || cast[0] != "Actor One" {
			t.Errorf("cast = %v", cast)
		}
		if len(months) != 2 {
			t.Errorf("months = %v, want 2 entries", months)
		}
	})

	t.Run("is_available filter", func(t *testing.T) {
		sites, tags, cast, months := queryFilters(db, boolPtr(true), nil)
		if len(sites) != 1 || sites[0] != "SiteA" {
			t.Errorf("sites = %v, want [SiteA]", sites)
		}
		if len(tags) != 2 {
			t.Errorf("tags = %v, want 2 entries", tags)
		}
		if len(cast) != 1 {
			t.Errorf("cast = %v, want 1 entry", cast)
		}
		if len(months) != 1 || months[0] != "2023-03" {
			t.Errorf("months = %v, want [2023-03]", months)
		}
	})

	t.Run("is_available false", func(t *testing.T) {
		sites, tags, cast, _ := queryFilters(db, boolPtr(false), nil)
		if len(sites) != 1 || sites[0] != "SiteB" {
			t.Errorf("sites = %v, want [SiteB]", sites)
		}
		if len(tags) != 0 || len(cast) != 0 {
			t.Errorf("expected no tags/cast for unavailable scenes, got %v %v", tags, cast)
		}
	})

	t.Run("soft-deleted scenes excluded", func(t *testing.T) {
		deletedTag := models.Tag{Name: "deleted-only-tag"}
		db.Create(&deletedTag)
		deletedScene := models.Scene{ID: 4, SceneID: "s4", Site: "DeletedSite", IsAvailable: true, IsAccessible: true, Tags: []models.Tag{deletedTag}}
		if err := db.Create(&deletedScene).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Delete(&deletedScene).Error; err != nil {
			t.Fatal(err)
		}

		sites, tags, _, _ := queryFilters(db, nil, nil)
		for _, s := range sites {
			if s == "DeletedSite" {
				t.Errorf("sites %v includes soft-deleted scene site", sites)
			}
		}
		for _, tag := range tags {
			if tag == "deleted-only-tag" {
				t.Errorf("tags %v includes soft-deleted scene tag", tags)
			}
		}
	})
}

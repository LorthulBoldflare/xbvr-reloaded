package models

import "github.com/jinzhu/gorm"

type DMSData struct {
	Sites        []string `json:"sites"`
	Actors       []string `json:"actors"`
	Tags         []string `json:"tags"`
	ReleaseGroup []string `json:"release_group"`
	Volumes      []Volume `json:"volumes"`
}

func GetDMSData() DMSData {
	db, _ := GetDB()

	// DISTINCT/aggregate queries instead of re-Find'ing full scene rows with
	// useless preloads for each facet
	base := func() *gorm.DB {
		// Table("scenes") bypasses gorm's soft-delete scope, so exclude
		// deleted scenes explicitly
		return db.Table("scenes").
			Where("scenes.deleted_at IS NULL").
			Where("scenes.is_accessible = ?", 1).
			Where("scenes.is_available = ?", 1)
	}

	// Available sites
	outSites := []string{}
	base().Where("scenes.site <> ''").Pluck("DISTINCT scenes.site", &outSites)

	// Available release dates (YYYY-MM)
	monthExpr := "strftime('%Y-%m', scenes.release_date)"
	if db.Dialect().GetName() == "mysql" {
		monthExpr = "DATE_FORMAT(scenes.release_date, '%Y-%m')"
	}
	outRelease := []string{}
	base().Pluck("DISTINCT "+monthExpr, &outRelease)

	// Available tags
	outTags := []string{}
	base().
		Joins("join scene_tags on scene_tags.scene_id = scenes.id").
		Joins("join tags on tags.id = scene_tags.tag_id").
		Where("tags.name <> ''").
		Order("tags.name asc").
		Pluck("DISTINCT tags.name", &outTags)

	// Available actors
	outCast := []string{}
	base().
		Joins("join scene_cast on scene_cast.scene_id = scenes.id").
		Joins("join actors on actors.id = scene_cast.actor_id").
		Where("actors.name <> ''").
		Order("actors.name asc").
		Pluck("DISTINCT actors.name", &outCast)

	// Available volumes
	var vol []Volume
	db.Where("is_available = ?", true).Find(&vol)

	return DMSData{Sites: outSites, Tags: outTags, Actors: outCast, Volumes: vol, ReleaseGroup: outRelease}
}

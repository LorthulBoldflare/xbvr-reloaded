package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/index/scorch"
	"github.com/sirupsen/logrus"
	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/models"
)

type Index struct {
	Bleve bleve.Index
}

type SceneIndexed struct {
	Description string    `json:"description"`
	Title       string    `json:"title"`
	Cast        string    `json:"cast"`
	Site        string    `json:"site"`
	Id          string    `json:"id"`
	Released    time.Time `json:"released"`
	Added       time.Time `json:"added"`
	Duration    int       `json:"duration"`
}

func NewIndex(name string) (*Index, error) {
	i := new(Index)

	path := filepath.Join(common.IndexDirV2, name)

	// the simple analyzer is more approriate for the title and cast
	// note this does not effect search unless the query includes cast: or title:
	titleFieldMapping := bleve.NewTextFieldMapping()
	titleFieldMapping.Analyzer = simple.Name
	castFieldMapping := bleve.NewTextFieldMapping()
	castFieldMapping.Analyzer = simple.Name
	releaseFieldMapping := bleve.NewDateTimeFieldMapping()
	addedFieldMapping := bleve.NewDateTimeFieldMapping()
	durationFieldMapping := bleve.NewNumericFieldMapping()
	sceneMapping := bleve.NewDocumentMapping()
	sceneMapping.AddFieldMappingsAt("title", titleFieldMapping)
	sceneMapping.AddFieldMappingsAt("cast", castFieldMapping)
	sceneMapping.AddFieldMappingsAt("released", releaseFieldMapping)
	sceneMapping.AddFieldMappingsAt("added", addedFieldMapping)
	sceneMapping.AddFieldMappingsAt("duration", durationFieldMapping)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("_default", sceneMapping)

	idx, err := bleve.NewUsing(path, mapping, scorch.Name, scorch.Name, nil)
	if err != nil && err == bleve.ErrorIndexPathExists {
		idx, err = bleve.Open(path)
	}
	if err != nil {
		return nil, err
	}

	i.Bleve = idx
	return i, nil
}

var (
	sharedIndexesMu sync.Mutex
	sharedIndexes   = map[string]*Index{}
)

// GetSharedIndex opens (once) and caches the bleve index for name. Bleve
// indexes are safe for concurrent use — callers must not close the handle.
// Use CloseSharedIndexes before deleting the index directory.
func GetSharedIndex(name string) (*Index, error) {
	sharedIndexesMu.Lock()
	defer sharedIndexesMu.Unlock()
	if idx, ok := sharedIndexes[name]; ok {
		return idx, nil
	}
	idx, err := NewIndex(name)
	if err != nil {
		return nil, err
	}
	sharedIndexes[name] = idx
	return idx, nil
}

// CloseSharedIndexes closes and forgets all cached index handles.
func CloseSharedIndexes() {
	sharedIndexesMu.Lock()
	defer sharedIndexesMu.Unlock()
	for name, idx := range sharedIndexes {
		idx.Bleve.Close()
		delete(sharedIndexes, name)
	}
}

// ResetSharedIndexes atomically closes all cached handles and deletes the
// index directory while holding the lock, so no concurrent search can cache
// a handle whose files are being removed.
func ResetSharedIndexes() error {
	sharedIndexesMu.Lock()
	defer sharedIndexesMu.Unlock()
	for name, idx := range sharedIndexes {
		idx.Bleve.Close()
		delete(sharedIndexes, name)
	}
	if err := os.RemoveAll(common.IndexDirV2); err != nil {
		return err
	}
	return os.MkdirAll(common.IndexDirV2, os.ModePerm)
}

// sceneIndexDoc builds the indexed document for a scene.
func sceneIndexDoc(scene models.Scene) SceneIndexed {
	cast := ""
	castConcat := ""
	for _, c := range scene.Cast {
		cast = cast + " " + c.Name
		castConcat = castConcat + " " + strings.Replace(c.Name, " ", "", -1)
	}

	rd := time.Date(scene.ReleaseDate.Year(), scene.ReleaseDate.Month(), scene.ReleaseDate.Day(), 0, 0, 0, 0, &time.Location{})
	return SceneIndexed{
		Title:       fmt.Sprintf("%v", scene.Title),
		Description: fmt.Sprintf("%v", scene.Synopsis),
		Cast:        fmt.Sprintf("%v %v", cast, castConcat),
		Site:        fmt.Sprintf("%v", scene.Site),
		Id:          fmt.Sprintf("%v", scene.SceneID),
		Released:    rd,                                       // only index the date, not the time
		Added:       scene.CreatedAt.Truncate(24 * time.Hour), // only index the date, not the time
		Duration:    scene.Duration,
	}
}

func SearchIndex() {
	if !models.CheckLock("index") {
		models.CreateLock("index")
		defer models.RemoveLock("index")

		tlog := log.WithFields(logrus.Fields{"task": "scrape"})

		// use the shared handle: a second open of the same scorch path blocks
		// forever on the bbolt file lock while the shared handle is held
		idx, err := GetSharedIndex("scenes")
		if err != nil {
			log.Error(err)
			models.RemoveLock("index")
			return
		}

		db, _ := models.GetDB()

		total := 0
		offset := 0
		current := 0
		var scenes []models.Scene
		tx := db.Model(models.Scene{}).Preload("Cast").Preload("Tags")
		tx.Count(&total)

		tlog.Infof("Building search index...")

		for {
			tx.Offset(offset).Limit(100).Find(&scenes)
			if len(scenes) == 0 {
				break
			}

			// batch-index the page (bleve.Batch is far cheaper than
			// per-document Index calls; indexing an existing ID overwrites)
			batch := idx.Bleve.NewBatch()
			for i := range scenes {
				batch.Index(scenes[i].SceneID, sceneIndexDoc(scenes[i]))
				current = current + 1
			}
			if err := idx.Bleve.Batch(batch); err != nil {
				log.Error(err)
			}
			tlog.Infof("Indexed %v/%v scenes", current, total)

			// Update migration status if migration is running
			if config.State.Migration.IsRunning {
				msg := fmt.Sprintf("Reindexing scenes: %v/%v", current, total)
				config.UpdateMigrationStatus(config.State.Migration.Current, current, total, msg)
			}

			offset = offset + 100
		}

		tlog.Infof("Search index built!")
	}
}

/**
 * Update search index for all of the specified scenes.
 */
func IndexScenes(scenes *[]models.Scene) {
	if !models.CheckLock("index") {
		models.CreateLock("index")
		defer models.RemoveLock("index")

		tlog := log.WithFields(logrus.Fields{"task": "scrape"})

		// use the shared handle: a second open of the same scorch path blocks
		// forever on the bbolt file lock while the shared handle is held
		idx, err := GetSharedIndex("scenes")
		if err != nil {
			log.Error(err)
			models.RemoveLock("index")
			return
		}

		tlog.Infof("Adding scraped scenes to search index...")

		total := 0
		lastMessage := time.Now()
		batch := idx.Bleve.NewBatch()
		for i := range *scenes {
			if time.Since(lastMessage) > time.Duration(config.Config.Advanced.ProgressTimeInterval)*time.Second {
				tlog.Infof("Indexed %v of %v scenes", total, len(*scenes))
				lastMessage = time.Now()
			}
			scene := (*scenes)[i]
			// indexing an existing ID overwrites the old document
			batch.Index(scene.SceneID, sceneIndexDoc(scene))
			total += 1
		}
		if err := idx.Bleve.Batch(batch); err != nil {
			log.Error(err)
		}

		tlog.Infof("Indexed %v scenes", total)
	}
}

func DeleteIndexScenes(scenes *[]models.Scene) {
	if !models.CheckLock("index") {
		models.CreateLock("index")
		defer models.RemoveLock("index")

		tlog := log.WithFields(logrus.Fields{"task": "scrape"})

		// use the shared handle: a second open of the same scorch path blocks
		// forever on the bbolt file lock while the shared handle is held
		idx, err := GetSharedIndex("scenes")
		if err != nil {
			log.Error(err)
			models.RemoveLock("index")
			return
		}

		tlog.Infof("Deleting scenes from search index...")

		total := 0
		lastMessage := time.Now()
		batch := idx.Bleve.NewBatch()
		for i := range *scenes {
			if time.Since(lastMessage) > time.Duration(config.Config.Advanced.ProgressTimeInterval)*time.Second {
				tlog.Infof("Deleting scene index %v of %v scenes", total, len(*scenes))
				lastMessage = time.Now()
			}
			scene := (*scenes)[i]
			batch.Delete(scene.SceneID)
			total += 1
		}
		if err := idx.Bleve.Batch(batch); err != nil {
			log.Error(err)
		}

		tlog.Infof("Deleted %v scenes from index", total)
	}
}

/**
 * Update search index for all of the specified scrapedScenes.
 * This method will first read the scraped scenes from the DB, after
 * which it calls IndexScenes.
 */
func IndexScrapedScenes(scrapedScenes *[]models.ScrapedScene) {
	// Map scrapedScenes to Scenes
	var scenes []models.Scene
	for i := range *scrapedScenes {
		var scene models.Scene
		scrapedScene := (*scrapedScenes)[i]
		// Read scraped scene from db, as we don't want to index it
		// if it doesn't exist in there
		err := scene.GetIfExist(scrapedScene.SceneID)
		if err == nil {
			scenes = append(scenes, scene)
		}
	}

	// Now update search index
	IndexScenes(&scenes)
}

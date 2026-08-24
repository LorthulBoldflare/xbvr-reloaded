package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/jinzhu/gorm"
	"github.com/markphelps/optional"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/ffprobe"
	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/scrape"
)

// The allowed video extensions are set in the config.go file as they now are user configurable
func getAllowedVideoExt() []string {
	if len(config.Config.Storage.VideoExt) == 0 {
		return config.DefaultVideoExtensions
	}
	return config.Config.Storage.VideoExt
}

func creationTime(fileTimes times.Timespec) time.Time {
	if fileTimes.HasBirthTime() {
		birthTime := fileTimes.BirthTime()
		if hasTimestamp(birthTime) {
			return birthTime
		}
	}

	return fileTimes.ModTime()
}

func hasTimestamp(timestamp time.Time) bool {
	return !timestamp.IsZero() && !timestamp.Equal(time.Unix(0, 0))
}

func needsCreationTimeRefresh(timestamp time.Time) bool {
	return !hasTimestamp(timestamp)
}

func RescanVolumes(id int) {
	if !models.CheckLock("rescan") {
		models.CreateLock("rescan")
		defer models.RemoveLock("rescan")

		tlog := log.WithFields(logrus.Fields{"task": "rescan"})
		tlog.Infof("Start scanning volumes")

		models.CheckVolumes()

		db, _ := models.GetDB()

		var vol []models.Volume
		if id > 0 {
			db.Where("id=?", id).Find(&vol)
		} else {
			db.Find(&vol)
		}

		var scannedLocalVolumeIDs []uint
		for i := range vol {
			tlog.Infof("Scanning %v", vol[i].Path)

			switch vol[i].Type {
			case "local":
				scanLocalVolume(vol[i], db, tlog)
				scannedLocalVolumeIDs = append(scannedLocalVolumeIDs, vol[i].ID)
			case "putio":
				scanPutIO(vol[i], db, tlog)
			}
		}

		if len(scannedLocalVolumeIDs) > 0 {
			var scannedScenes []models.Scene
			db.Model(&models.Scene{}).
				Joins("JOIN files ON files.scene_id = scenes.id").
				Where("files.volume_id IN (?) AND files.scene_id != 0", scannedLocalVolumeIDs).
				Group("scenes.id").
				Find(&scannedScenes)

			for i := range scannedScenes {
				scannedScenes[i].UpdateStatus()
			}
		}

		// Match Scene to File
		var files []models.File

		tlog.Infof("Matching Scenes to known filenames")
		db.Model(&models.File{}).Where("files.scene_id = 0").Find(&files)

		escape := func(s string) string {
			var buffer bytes.Buffer
			json.HTMLEscape(&buffer, []byte(s))
			return buffer.String()
		}

		// filename variants tried for each unmatched file
		filenameVariants := func(filename string) []string {
			return []string{
				filename,
				strings.Replace(filename, ".funscript", ".mp4", -1),
				strings.Replace(filename, ".hsp", ".mp4", -1),
				strings.Replace(filename, ".srt", ".mp4", -1),
				strings.Replace(filename, ".cmscript", ".mp4", -1),
			}
		}

		// Load all scene filename lists once and match in memory — the
		// previous per-file LIKE '%..%' queries were O(unmatched × scenes)
		// full table scans.
		var sceneRows []struct {
			ID           uint
			FilenamesArr string
		}
		// deleted_at filter: scanning into an anonymous struct bypasses
		// gorm's soft-delete scope, so exclude deleted scenes explicitly —
		// otherwise files bind to deleted scenes (and UpdateStatus on the
		// failed lookup would insert a blank scene row)
		db.Table("scenes").Select("id, filenames_arr").Where("scenes.deleted_at IS NULL").Scan(&sceneRows)
		sceneByFilename := map[string][]uint{}
		for _, row := range sceneRows {
			var names []string
			if err := json.Unmarshal([]byte(row.FilenamesArr), &names); err != nil {
				continue
			}
			for _, name := range names {
				// lowercase keys: the previous LIKE-based matching was
				// case-insensitive for ASCII on sqlite and default MySQL
				// collations
				key := strings.ToLower(escape(name))
				sceneByFilename[key] = append(sceneByFilename[key], row.ID)
			}
		}

		// Same for alternate-source external references when enabled
		var extrefs []models.ExternalReference
		var extrefDataLower []string
		if config.Config.Advanced.UseAltSrcInFileMatching {
			db.Preload("XbvrLinks").Where("external_source like 'alternate scene %'").Find(&extrefs)
			// lowercase once: the previous LIKE-based matching was
			// case-insensitive for ASCII
			extrefDataLower = make([]string, len(extrefs))
			for i, extref := range extrefs {
				extrefDataLower[i] = strings.ToLower(extref.ExternalData)
			}
		}

		for i := range files {
			filename := strings.ToLower(escape(path.Base(files[i].Filename)))
			variants := filenameVariants(filename)

			matchedSceneIDs := map[uint]bool{}
			for _, variant := range variants {
				for _, id := range sceneByFilename[variant] {
					matchedSceneIDs[id] = true
				}
			}

			if len(matchedSceneIDs) == 1 {
				for id := range matchedSceneIDs {
					files[i].SceneID = id
				}
				files[i].Save()
				var scene models.Scene
				if err := scene.GetIfExistByPK(files[i].SceneID); err == nil {
					scene.UpdateStatus()
				}
			} else {
				if len(matchedSceneIDs) == 0 && config.Config.Advanced.UseAltSrcInFileMatching {
					// check if the filename matches in an external_reference record
					var matchedExtRefs []models.ExternalReference
					for i, extref := range extrefs {
						for _, variant := range variants {
							if strings.Contains(extrefDataLower[i], `"`+variant) {
								matchedExtRefs = append(matchedExtRefs, extref)
								break
							}
						}
					}
					if len(matchedExtRefs) == 1 && len(matchedExtRefs[0].XbvrLinks) == 1 {
						// the scene id will be the Internal DB Id from the associated link
						var scene models.Scene
						if err := scene.GetIfExistByPK(matchedExtRefs[0].XbvrLinks[0].InternalDbId); err != nil {
							// dangling link (e.g. soft-deleted scene) — skip instead
							// of saving a zero-value scene, which inserts a blank row
							continue
						}
						// Add File to the list of Scene filenames
						var pfTxt []string
						if err := json.Unmarshal([]byte(scene.FilenamesArr), &pfTxt); err != nil {
							continue
						}
						pfTxt = append(pfTxt, files[i].Filename)
						if tmp, err := json.Marshal(pfTxt); err == nil {
							scene.FilenamesArr = string(tmp)
						}
						scene.Save()

						files[i].SceneID = scene.ID
						files[i].Save()
						scene.UpdateStatus()
					}
				}
				if files[i].SceneID == 0 && config.Config.Storage.MatchOhash && config.Config.Advanced.StashApiKey != "" {
					hash := files[i].OsHash
					if len(hash) < 16 {
						// the has in xbvr is sometiomes < 16 pad with zeros
						paddingLength := 16 - len(hash)
						hash = strings.Repeat("0", paddingLength) + hash
					}
					queryVariable := scrape.StashVariablesJSON(map[string]interface{}{
						"fingerprints": map[string]interface{}{
							"value":    hash,
							"modifier": "INCLUDES",
						},
						"page": 1,
					})
					// call Stashdb graphql searching for os_hash
					stashMatches := scrape.GetScenePage(queryVariable)
					for _, match := range stashMatches.Data.QueryScenes.Scenes {
						if match.ID != "" {
							var externalRefLink models.ExternalReferenceLink
							db.Where(&models.ExternalReferenceLink{ExternalSource: "stashdb scene", ExternalId: match.ID}).First(&externalRefLink)
							if externalRefLink.ID != 0 {
								var scene models.Scene
								if err := scene.GetIfExistByPK(externalRefLink.InternalDbId); err != nil {
									// dangling link (e.g. soft-deleted scene) — skip
									// instead of binding the file to it and saving a
									// zero-value scene, which inserts a blank row
									continue
								}
								files[i].SceneID = externalRefLink.InternalDbId
								files[i].Save()

								// add filename tyo the array
								var pfTxt []string
								json.Unmarshal([]byte(scene.FilenamesArr), &pfTxt)
								pfTxt = append(pfTxt, files[i].Filename)
								tmp, _ := json.Marshal(pfTxt)
								scene.FilenamesArr = string(tmp)
								scene.Save()
								models.AddAction(scene.SceneID, "match", "filenames_arr", scene.FilenamesArr)

								scene.UpdateStatus()
								log.Infof("File %s matched to Scene %s matched using stashdb hash %s", path.Base(files[i].Filename), scene.SceneID, hash)
							}
						}
					}
				}
			}

			if (i % 50) == 0 {
				tlog.Infof("Matching Scenes to known filenames (%v/%v)", i+1, len(files))
			}
		}

		tlog.Infof("Generating heatmaps")

		GenerateHeatmaps(tlog)

		tlog.Infof("Scanning complete")

		// Inform UI about state change
		common.PublishWS("state.change.optionsStorage", nil)

		// Grab metrics
		var localFilesCount int64
		var localFilesSize int64
		// single pass: count and sum sizes in SQL instead of scanning all rows
		db.Model(models.File{}).
			Joins("left join volumes on files.volume_id = volumes.id").
			Where("volumes.type = ?", "local").
			Select("count(*) as cnt, coalesce(sum(files.size), 0) as size").
			Row().Scan(&localFilesCount, &localFilesSize)
		common.AddMetricPoint("local_files_count", float64(localFilesCount))
		common.AddMetricPoint("local_files_size", float64(localFilesSize))

		r := models.RequestSceneList{}
		common.AddMetricPoint("scenes_scraped", float64(models.QueryScenes(r, false).Results))

		r = models.RequestSceneList{IsAvailable: optional.NewBool(true)}
		common.AddMetricPoint("scenes_downloaded", float64(models.QueryScenes(r, false).Results))

		r = models.RequestSceneList{IsWatched: optional.NewBool(true)}
		common.AddMetricPoint("scenes_watched_overall", float64(models.QueryScenes(r, false).Results))

		r = models.RequestSceneList{IsWatched: optional.NewBool(false), IsAvailable: optional.NewBool(true)}
		common.AddMetricPoint("scenes_downloaded_unwatched", float64(models.QueryScenes(r, false).Results))
	}
}

func scanLocalVolume(vol models.Volume, db *gorm.DB, tlog *logrus.Entry) {
	allowedVideoExt := getAllowedVideoExt()
	if vol.IsMounted() {

		var videoProcList []string
		var scriptProcList []string
		var hspProcList []string
		var subtitlesProcList []string
		_ = filepath.Walk(vol.Path, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !f.Mode().IsDir() {
				// Make sure the filename should be considered
				if !strings.HasPrefix(filepath.Base(path), ".") && funk.Contains(allowedVideoExt, strings.ToLower(filepath.Ext(path))) {
					var fl models.File
					err = db.Where(&models.File{Path: filepath.Dir(path), Filename: filepath.Base(path)}).First(&fl).Error

					if err == gorm.ErrRecordNotFound || fl.VolumeID == 0 || fl.VideoDuration == 0 || fl.VideoProjection == "" || fl.Size != f.Size() || fl.OsHash == "" || needsCreationTimeRefresh(fl.CreatedTime) {
						videoProcList = append(videoProcList, path)
					}
				}

				if !strings.HasPrefix(filepath.Base(path), ".") && (filepath.Ext(path) == ".funscript" || strings.ToLower(filepath.Ext(path)) == ".cmscript") {
					scriptProcList = append(scriptProcList, path)
				}
				if !strings.HasPrefix(filepath.Base(path), ".") && filepath.Ext(path) == ".hsp" {
					hspProcList = append(hspProcList, path)
				}
				if !strings.HasPrefix(filepath.Base(path), ".") && (filepath.Ext(path) == ".srt" || filepath.Ext(path) == ".ssa" || filepath.Ext(path) == ".ass") {
					subtitlesProcList = append(subtitlesProcList, path)
				}
			}
			return nil
		})

		filenameSeparator := regexp.MustCompile("[ _.-]+")

		for j, path := range videoProcList {
			fStat, err := os.Stat(path)
			if err != nil {
				tlog.Errorf("Can't stat %s, error: %s", path, err)
				continue
			}
			fTimes, err := times.Stat(path)
			if err != nil {
				tlog.Errorf("Can't get the modification/creation times for %s, error: %s", path, err)
				continue
			}

			var fl models.File
			db.Where(&models.File{
				Path:     filepath.Dir(path),
				Filename: filepath.Base(path),
				Type:     "video",
			}).FirstOrCreate(&fl)

			fl.Size = fStat.Size()
			fl.CreatedTime = creationTime(fTimes)
			fl.UpdatedTime = fTimes.ModTime()
			fl.VolumeID = vol.ID

			hash, err := Hash(path)
			if err == nil {
				fl.OsHash = fmt.Sprintf("%x", hash)
			}

			ffdata, err := ffprobe.GetProbeData(path, time.Second*5)
			if err != nil {
				tlog.Error("Error running ffprobe", path, err)
			} else {
				vs := ffdata.GetFirstVideoStream()
				if vs == nil {
					tlog.Error("No video stream in file ", path)
				} else {
					if vs.BitRate != "" {
						bitRate, _ := strconv.Atoi(vs.BitRate)
						fl.VideoBitRate = bitRate
					}
					fl.VideoAvgFrameRate = vs.AvgFrameRate
					fl.VideoCodecName = vs.CodecName
					fl.VideoWidth = vs.Width
					fl.VideoHeight = vs.Height
					if dur, err := strconv.ParseFloat(vs.Duration, 64); err == nil {
						fl.VideoDuration = dur
					} else if ffdata.Format.DurationSeconds > 0.0 {
						fl.VideoDuration = ffdata.Format.DurationSeconds
					}
					fl.HasAlpha = false

					if vs.Height*2 == vs.Width || vs.Width > vs.Height {
						fl.VideoProjection = "180_sbs"
						nameparts := filenameSeparator.Split(strings.ToLower(filepath.Base(path)), -1)
						for i, part := range nameparts {
							if part == "mkx200" || part == "mkx220" || part == "rf52" || part == "fisheye190" || part == "vrca220" || part == "flat" {
								fl.VideoProjection = part
								break
							} else if part == "fisheye" || part == "f180" || part == "180f" {
								fl.VideoProjection = "fisheye"
								break
							} else if i < len(nameparts)-1 && (part+"_"+nameparts[i+1] == "mono_360" || part+"_"+nameparts[i+1] == "mono_180") {
								fl.VideoProjection = nameparts[i+1] + "_mono"
								break
							} else if i < len(nameparts)-1 && (part+"_"+nameparts[i+1] == "360_mono" || part+"_"+nameparts[i+1] == "180_mono") {
								fl.VideoProjection = part + "_mono"
								break
							}
						}
						if fl.VideoProjection == "mkx200" || fl.VideoProjection == "mkx220" || fl.VideoProjection == "rf52" || fl.VideoProjection == "fisheye190" || fl.VideoProjection == "vrca220" {
							// alpha passthrough only works with fisheye projections
							for _, part := range nameparts {
								if part == "alpha" {
									fl.HasAlpha = true
									break
								}
							}
						}
					}

					if vs.Height == vs.Width {
						fl.VideoProjection = "360_tb"
					}

					fl.CalculateFramerate()
				}
			}

			err = fl.Save()
			if err != nil {
				tlog.Errorf("New file %s, but got error %s", path, err)
			}

			tlog.Infof("Scanning %v (%v/%v)", vol.Path, j+1, len(videoProcList))
		}

		for _, path := range scriptProcList {
			var fl models.File
			db.Where(&models.File{
				Path:     filepath.Dir(path),
				Filename: filepath.Base(path),
				Type:     "script",
			}).FirstOrCreate(&fl)

			fStat, err := os.Stat(path)
			if err != nil {
				tlog.Errorf("Can't stat %s, error: %s", path, err)
				continue
			}
			fTimes, err := times.Stat(path)
			if err != nil {
				tlog.Errorf("Can't get the modification/creation times for %s, error: %s", path, err)
				continue
			}

			if fStat.Size() != fl.Size {
				fl.Size = fStat.Size()
				fl.HasHeatmap = false
				fl.VideoDuration = 0.0
			}

			if fl.VideoDuration < 0.01 {
				duration, err := getFunscriptDuration(path)
				if err == nil {
					fl.VideoDuration = duration
				}
			}

			fl.CreatedTime = creationTime(fTimes)
			fl.UpdatedTime = fTimes.ModTime()
			fl.VolumeID = vol.ID
			fl.Save()
		}

		for _, path := range hspProcList {
			ScanLocalHspFile(path, vol.ID, 0)
		}

		for _, path := range subtitlesProcList {
			ScanLocalSubtitlesFile(path, vol.ID, 0)
		}

		vol.LastScan = time.Now()
		vol.Save()

		var scene models.Scene
		// Check if files are still present at the location
		allFiles := vol.Files()
		for i := range allFiles {
			if !allFiles[i].Exists() {
				log.Info(allFiles[i].GetPath())
				db.Delete(&allFiles[i])
				if allFiles[i].SceneID != 0 {
					scene.GetIfExistByPK(allFiles[i].SceneID)
					scene.UpdateStatus()
				}
			}
		}
	}
}

func scanPutIO(vol models.Volume, db *gorm.DB, tlog *logrus.Entry) {
	allowedVideoExt := getAllowedVideoExt()
	client := vol.GetPutIOClient()

	acct, err := client.Account.Info(context.Background())
	if err != nil {
		vol.IsAvailable = false
		vol.Save()
		return
	}

	files, _, err := client.Files.List(context.Background(), -1)
	if err != nil {
		return
	}

	// Walk
	var currentFileID []string
	for i := range files {
		if !files[i].IsDir() && funk.Contains(allowedVideoExt, strings.ToLower(filepath.Ext(files[i].Name))) {
			var fl models.File
			err = db.Where(&models.File{Path: strconv.FormatInt(files[i].ID, 10), Filename: files[i].Name}).First(&fl).Error

			if err == gorm.ErrRecordNotFound {
				var fl models.File
				db.Where(&models.File{
					Path:     strconv.FormatInt(files[i].ID, 10),
					Filename: files[i].Name,
				}).FirstOrCreate(&fl)
				fl.VideoProjection = "180_sbs"
				fl.Size = files[i].Size
				fl.Type = "video"
				fl.CreatedTime = files[i].CreatedAt.Time
				fl.UpdatedTime = files[i].UpdatedAt.Time
				fl.VolumeID = vol.ID
				fl.OsHash = files[i].OpensubtitlesHash
				fl.Save()
			}

			currentFileID = append(currentFileID, strconv.FormatInt(files[i].ID, 10))
		}
	}

	var scene models.Scene
	// Check if local files are present in listing
	allFiles := vol.Files()
	for i := range allFiles {
		if !funk.ContainsString(currentFileID, allFiles[i].Path) {
			log.Info(allFiles[i].GetPath())
			db.Delete(&allFiles[i])
			if allFiles[i].SceneID != 0 {
				scene.GetIfExistByPK(allFiles[i].SceneID)
				scene.UpdateStatus()
			}
		}
	}

	// Update volume info
	vol.IsAvailable = true
	vol.Path = "Put.io (" + acct.Username + ")"
	vol.LastScan = time.Now()
	vol.Save()
}
func RefreshSceneStatuses() {
	// refreshes the status of all scenes
	tlog := log.WithFields(logrus.Fields{"task": "rescan"})
	tlog.Infof("Update status of Scenes")
	db, _ := models.GetDB()

	// iterate IDs and load each scene individually instead of materializing
	// the entire scenes table at once
	var ids []uint
	db.Model(&models.Scene{}).Pluck("id", &ids)

	for i, id := range ids {
		var scene models.Scene
		if err := db.First(&scene, id).Error; err == nil {
			scene.UpdateStatus()
		}
		if (i % 70) == 0 {
			tlog.Infof("Update status of Scenes (%v/%v)", i+1, len(ids))
		}
	}

	tlog.Infof("Scene status refresh complete")
}
func ScanLocalHspFile(path string, volID uint, sceneId uint) {
	db, _ := models.GetDB()

	var fl models.File
	db.Where(&models.File{
		Path:     filepath.Dir(path),
		Filename: filepath.Base(path),
		Type:     "hsp",
	}).FirstOrCreate(&fl)

	fStat, err := os.Stat(path)
	if err != nil {
		log.Errorf("Can't stat %s, error: %s", path, err)
		return
	}
	fTimes, err := times.Stat(path)
	if err != nil {
		log.Errorf("Can't get the modification/creation times for %s, error: %s", path, err)
		return
	}

	fl.Size = fStat.Size()
	fl.CreatedTime = creationTime(fTimes)
	fl.UpdatedTime = fTimes.ModTime()
	fl.VolumeID = volID
	if sceneId > 0 {
		fl.SceneID = sceneId
	}
	fl.Save()

}

func ScanLocalSubtitlesFile(path string, volID uint, sceneId uint) {
	db, _ := models.GetDB()

	var fl models.File
	db.Where(&models.File{
		Path:     filepath.Dir(path),
		Filename: filepath.Base(path),
		Type:     "subtitles",
	}).FirstOrCreate(&fl)

	fStat, err := os.Stat(path)
	if err != nil {
		log.Errorf("Can't stat %s, error: %s", path, err)
		return
	}
	fTimes, err := times.Stat(path)
	if err != nil {
		log.Errorf("Can't get the modification/creation times for %s, error: %s", path, err)
		return
	}

	fl.Size = fStat.Size()
	fl.CreatedTime = creationTime(fTimes)
	fl.UpdatedTime = fTimes.ModTime()
	fl.VolumeID = volID
	if sceneId > 0 {
		fl.SceneID = sceneId
	}
	fl.Save()
}

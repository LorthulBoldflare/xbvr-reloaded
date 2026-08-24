package tasks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/ffprobe"
	"github.com/xbapps/xbvr/pkg/models"
)

// PreviewQueueStatus describes the state of the preview generation queue and
// is propagated to the UI via the "options.previews.queue" websocket topic.
type PreviewQueueStatus struct {
	Running      bool   `json:"running"`
	Stopping     bool   `json:"stopping"`
	Total        int    `json:"total"`
	Completed    int    `json:"completed"`
	Remaining    int    `json:"remaining"`
	CurrentScene string `json:"currentScene"`
}

var previewQueue = struct {
	sync.Mutex
	status PreviewQueueStatus
	stop   bool
}{}

// GetPreviewQueueStatus returns a snapshot of the preview generation queue.
func GetPreviewQueueStatus() PreviewQueueStatus {
	previewQueue.Lock()
	defer previewQueue.Unlock()
	return previewQueue.status
}

// StopPreviewGeneration requests the preview generation queue to stop. The
// currently running ffmpeg process is killed (within ~500ms) and no further
// snippets or scenes are processed.
func StopPreviewGeneration() {
	previewQueue.Lock()
	if previewQueue.status.Running {
		previewQueue.stop = true
		previewQueue.status.Stopping = true
	}
	previewQueue.Unlock()
	publishPreviewQueueStatus()
}

func previewStopRequested() bool {
	previewQueue.Lock()
	defer previewQueue.Unlock()
	return previewQueue.stop
}

func publishPreviewQueueStatus() {
	previewQueue.Lock()
	status := previewQueue.status
	previewQueue.Unlock()

	common.PublishWS("options.previews.queue", map[string]interface{}{
		"running":      status.Running,
		"stopping":     status.Stopping,
		"total":        status.Total,
		"completed":    status.Completed,
		"remaining":    status.Remaining,
		"currentScene": status.CurrentScene,
	})
}

func GeneratePreviews(endTime *time.Time) {
	if models.CheckLock("previews") {
		return
	}
	models.CreateLock("previews")
	defer models.RemoveLock("previews")

	log.Infof("Generating previews")
	db, _ := models.GetDB()

	var scenes []models.Scene
	db.Model(&models.Scene{}).Where("is_available = ?", true).Where("has_video_preview = ?", false).Order("release_date desc").Find(&scenes)

	previewQueue.Lock()
	previewQueue.stop = false
	previewQueue.status = PreviewQueueStatus{
		Running:   true,
		Total:     len(scenes),
		Remaining: len(scenes),
	}
	previewQueue.Unlock()
	publishPreviewQueueStatus()

	defer func() {
		previewQueue.Lock()
		previewQueue.status.Running = false
		previewQueue.status.Stopping = false
		previewQueue.status.CurrentScene = ""
		previewQueue.Unlock()
		publishPreviewQueueStatus()
	}()

	for _, scene := range scenes {
		if previewStopRequested() {
			log.Infof("Preview generation stopped by user")
			break
		}
		files, _ := scene.GetFiles()
		if len(files) > 0 {
			if endTime != nil && time.Now().After(*endTime) {
				return
			}
			previewQueue.Lock()
			previewQueue.status.CurrentScene = scene.SceneID
			previewQueue.Unlock()
			publishPreviewQueueStatus()

			i := 0
			for i < len(files) && files[i].Exists() {
				if previewStopRequested() {
					break
				}
				if files[i].Type == "video" {
					log.Infof("Rendering %v", scene.SceneID)
					destFile := filepath.Join(common.VideoPreviewDir, scene.SceneID+".mp4")
					err := RenderPreview(
						files[i].GetPath(),
						destFile,
						files[i].VideoProjection,
						config.Config.Library.Preview.StartTime,
						config.Config.Library.Preview.SnippetLength,
						config.Config.Library.Preview.SnippetAmount,
						config.Config.Library.Preview.Resolution,
						config.Config.Library.Preview.ExtraSnippet,
					)
					if err == nil {
						scene.HasVideoPreview = true
						scene.Save()
						break
					} else {
						log.Warn(err)
					}
				}
				i++
			}
		}

		previewQueue.Lock()
		previewQueue.status.Completed++
		previewQueue.status.Remaining--
		previewQueue.Unlock()
		publishPreviewQueueStatus()
	}
	log.Infof("Previews generated")
}

func RenderPreview(inputFile string, destFile string, videoProjection string, startTime int, snippetLength float64, snippetAmount int, resolution int, extraSnippet bool) error {
	// Use a per-render temp directory: a shared one races when previews are
	// rendered concurrently (e.g. queue worker + test-preview from options).
	if err := os.MkdirAll(common.VideoPreviewDir, os.ModePerm); err != nil {
		return err
	}
	tmpPath, err := os.MkdirTemp(common.VideoPreviewDir, "tmp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpPath)

	ffmpegPath, err := resolveBinary("ffmpeg")
	if err != nil {
		return err
	}

	// Get video duration
	ffdata, err := ffprobe.GetProbeData(inputFile, time.Second*10)
	if err != nil {
		return err
	}
	vs := ffdata.GetFirstVideoStream()
	if vs == nil {
		return errors.New("no video stream found")
	}
	dur := ffdata.Format.DurationSeconds

	crop := "iw/2:ih:iw/2:ih" // LR videos
	rect := cropRect{w: vs.Width / 2, h: vs.Height, x: vs.Width / 2, y: 0}
	if vs.Height == vs.Width {
		crop = "iw/2:ih/2:iw/4:ih/2" // TB videos
		rect = cropRect{w: vs.Width / 2, h: vs.Height / 2, x: vs.Width / 4, y: vs.Height / 2}
	}
	flat := videoProjection == "flat"
	if flat {
		crop = "iw:ih:iw:ih" // flat videos
		rect = cropRect{w: vs.Width, h: vs.Height, x: 0, y: 0}
	}
	// Mono 360 crop args: (no way of accurately determining)
	// "iw/2:ih:iw/4:ih"

	transcoder, err := newPreviewTranscoder(
		ffmpegPath, inputFile, vs.CodecName, vs.PixFmt, vs.Width, vs.Height, flat, crop, rect, resolution, snippetLength, startTime,
	)
	if err != nil {
		return err
	}

	// Prepare snippets
	interval := (dur - float64(startTime)) / float64(snippetAmount)
	for i := 1; i <= snippetAmount; i++ {
		if previewStopRequested() {
			return errors.New("preview generation stopped")
		}
		start := time.Duration(float64(i)*interval+float64(startTime)) * time.Second
		snippetFile := filepath.Join(tmpPath, fmt.Sprintf("%v.mp4", i))
		if err := transcoder.renderSnippet(start, snippetFile); err != nil {
			return err
		}
	}

	// Ensure ending is always in preview
	if extraSnippet && dur/float64(snippetAmount) > float64(150) && !previewStopRequested() {
		snippetAmount = snippetAmount + 1

		start := time.Duration(dur-float64(150)) * time.Second
		snippetFile := filepath.Join(tmpPath, fmt.Sprintf("%v.mp4", snippetAmount))
		if err := transcoder.renderSnippet(start, snippetFile); err != nil {
			return err
		}
	}

	// Prepare concat file
	concatFile := filepath.Join(tmpPath, "concat.txt")
	f, err := os.Create(concatFile)
	if err != nil {
		return err
	}
	for i := 1; i <= snippetAmount; i++ {
		f.WriteString(fmt.Sprintf("file '%v.mp4'\n", i))
	}
	f.Close()

	// Save result
	if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
		return err
	}
	args := []string{
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", filepath.ToSlash(concatFile),
		"-c", "copy",
		filepath.ToSlash(destFile),
	}
	if err := runFFmpeg(ffmpegPath, args...); err != nil {
		return err
	}

	return nil
}

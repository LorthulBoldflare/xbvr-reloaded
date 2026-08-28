package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/jinzhu/gorm"

	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/models"
)

// setupDeeplinkDB seeds an in-memory database with one scene and one video
// file, then swaps it in as the shared models DB for the test.
func setupDeeplinkDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open("sqlite3", "file:"+strings.ReplaceAll(t.Name(), "/", "-")+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	db.DB().SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&models.Scene{}, &models.File{}, &models.SceneCuepoint{},
		&models.Tag{}, &models.Actor{}, &models.History{}, &models.Volume{},
	).Error; err != nil {
		t.Fatal(err)
	}

	scene := models.Scene{
		ID:       123,
		SceneID:  "scene-123",
		Title:    "Test Scene",
		Site:     "TestSite",
		CoverURL: "https://cdn.example.com/cover.jpg",
	}
	if err := db.Create(&scene).Error; err != nil {
		t.Fatal(err)
	}
	file := models.File{
		SceneID:         123,
		Type:            "video",
		Filename:        "test.mp4",
		Size:            1024,
		VideoHeight:     2160,
		VideoWidth:      4320,
		VideoDuration:   120,
		VideoProjection: "180_sbs",
	}
	if err := db.Create(&file).Error; err != nil {
		t.Fatal(err)
	}

	restore := models.SetCommonDBForTests(db)
	t.Cleanup(func() {
		restore()
		db.Close()
	})
}

func setupDeeplinkConfig(t *testing.T, enabled bool, publicURL string) {
	t.Helper()
	origDeo := config.Config.Interfaces.DeoVR
	origPublic := config.Config.Server.PublicURL
	t.Cleanup(func() {
		config.Config.Interfaces.DeoVR = origDeo
		config.Config.Server.PublicURL = origPublic
	})
	config.Config.Interfaces.DeoVR.Enabled = enabled
	config.Config.Server.PublicURL = publicURL
}

func dispatchDeeplinkResource(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	container := restful.NewContainer()
	container.Add(DeoVRDeeplinkResource{}.WebService())

	rec := httptest.NewRecorder()
	container.Dispatch(rec, req)
	return rec
}

func TestDeoVRDeeplinkEndpoint(t *testing.T) {
	setupDeeplinkDB(t)
	setupDeeplinkConfig(t, true, "https://my.xbvr.reloaded")

	deeplinkReq := func(path string) *http.Request {
		return httptest.NewRequest(http.MethodGet, "http://xbvr.local"+path, nil)
	}

	t.Run("serves single-video payload with public URL base", func(t *testing.T) {
		rec := dispatchDeeplinkResource(t, deeplinkReq("/api/deovr/123.json"))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}

		var payload DeoScene
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("response is not a DeoScene payload: %v", err)
		}
		if payload.ID != 123 || payload.Title != "Test Scene" {
			t.Fatalf("unexpected payload identity: id=%d title=%q", payload.ID, payload.Title)
		}
		if !payload.Is3D || payload.StereoMode != "sbs" || payload.ScreenType != "dome" {
			t.Fatalf("unexpected projection fields: is3d=%v stereoMode=%q screenType=%q",
				payload.Is3D, payload.StereoMode, payload.ScreenType)
		}
		if payload.VideoLength != 120 {
			t.Fatalf("VideoLength = %d, want 120", payload.VideoLength)
		}
		if len(payload.Encodings) != 1 || len(payload.Encodings[0].VideoSources) != 1 {
			t.Fatalf("expected one encoding with one video source, got %+v", payload.Encodings)
		}
		wantPrefix := "https://my.xbvr.reloaded/api/dms/file/"
		if got := payload.Encodings[0].VideoSources[0].URL; !strings.HasPrefix(got, wantPrefix) {
			t.Fatalf("video source URL = %q, want prefix %q", got, wantPrefix)
		}
	})

	t.Run("falls back to request host without public URL", func(t *testing.T) {
		config.Config.Server.PublicURL = ""
		defer func() { config.Config.Server.PublicURL = "https://my.xbvr.reloaded" }()

		rec := dispatchDeeplinkResource(t, deeplinkReq("/api/deovr/123.json"))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var payload DeoScene
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if got := payload.Encodings[0].VideoSources[0].URL; !strings.HasPrefix(got, "http://xbvr.local/api/dms/file/") {
			t.Fatalf("video source URL = %q, want request-host prefix", got)
		}
	})

	t.Run("requires .json suffix", func(t *testing.T) {
		rec := dispatchDeeplinkResource(t, deeplinkReq("/api/deovr/123"))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 without .json suffix, got %d", rec.Code)
		}
	})

	t.Run("unknown scene is 404", func(t *testing.T) {
		rec := dispatchDeeplinkResource(t, deeplinkReq("/api/deovr/999.json"))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for unknown scene, got %d", rec.Code)
		}
	})

	t.Run("disabled DeoVR interface is 404", func(t *testing.T) {
		config.Config.Interfaces.DeoVR.Enabled = false
		defer func() { config.Config.Interfaces.DeoVR.Enabled = true }()

		rec := dispatchDeeplinkResource(t, deeplinkReq("/api/deovr/123.json"))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 while DeoVR interface disabled, got %d", rec.Code)
		}
	})
}

// The scene detail payload carries the per-scene deeplink token so the web UI
// can build the "Open in DeoVR" link; it is omitted when player auth is off.
func TestSceneDetailIncludesDeeplinkToken(t *testing.T) {
	setupDeeplinkDB(t)
	setupPlayerAuth(t, true)

	dispatch := func(t *testing.T, path string) *httptest.ResponseRecorder {
		t.Helper()
		container := restful.NewContainer()
		container.Add(SceneResource{}.WebService())
		rec := httptest.NewRecorder()
		container.Dispatch(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}

	t.Run("token present when player auth enabled", func(t *testing.T) {
		rec := dispatch(t, "/api/scene/123")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
		}
		var scene models.Scene
		if err := json.Unmarshal(rec.Body.Bytes(), &scene); err != nil {
			t.Fatal(err)
		}
		if scene.DeoVRDeeplinkToken != config.DeoVRDeeplinkToken(123) {
			t.Fatalf("deovr_deeplink_token = %q, want %q", scene.DeoVRDeeplinkToken, config.DeoVRDeeplinkToken(123))
		}
	})

	t.Run("token omitted when player auth disabled", func(t *testing.T) {
		config.Config.Interfaces.DeoVR.AuthEnabled = false
		defer func() { config.Config.Interfaces.DeoVR.AuthEnabled = true }()

		rec := dispatch(t, "/api/scene/123")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var raw map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}
		if _, present := raw["deovr_deeplink_token"]; present {
			t.Fatal("deovr_deeplink_token must be omitted while player auth disabled")
		}
	})
}

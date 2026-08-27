package server

import (
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jinzhu/gorm"
	"willnorris.com/go/imageproxy"

	"github.com/xbapps/xbvr/pkg/models"
)

// forceCacheHeaderTransport mimics NewForceCacheTransport's cache-header
// forcing, without its SSRF transport (which would block the loopback test
// server). Required because httpcache only stores responses it deems
// cacheable — the forced Cache-Control header is what makes upstream 2xx
// responses storable in the disk cache.
type forceCacheHeaderTransport struct{}

func (forceCacheHeaderTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		resp.Header.Set("Cache-Control", "public, max-age=157680000")
	}
	return resp, nil
}

type ctxTestUpstream struct {
	server  *httptest.Server
	hits    *int64
	lastURI atomic.Value
}

func newCtxTestUpstream(t *testing.T) *ctxTestUpstream {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	u := &ctxTestUpstream{hits: new(int64)}
	u.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(u.hits, 1)
		u.lastURI.Store(r.RequestURI)
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, img, nil)
	}))
	t.Cleanup(u.server.Close)
	return u
}

func newCtxTestHandler(t *testing.T, upstream *ctxTestUpstream, cacheable bool) *ImageProxyContextHandler {
	t.Helper()
	var transport http.RoundTripper
	if cacheable {
		transport = forceCacheHeaderTransport{}
	}
	return NewImageProxyContextHandler(imageproxy.NewProxy(transport, diskCache(t.TempDir())))
}

func doImgRequest(h http.Handler, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func setupImgCtxDB(t *testing.T) (*gorm.DB, models.Scene, models.Actor) {
	t.Helper()
	db, err := models.GetCommonDB()
	if err != nil {
		t.Fatalf("get db: %v", err)
	}
	if err := db.AutoMigrate(&models.Scene{}, &models.Actor{}, &models.ImageProxyEntry{}).Error; err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	scene := models.Scene{SceneID: "imgctx-test-scene-1", Title: "imgctx test"}
	if err := db.Create(&scene).Error; err != nil {
		t.Fatal(err)
	}
	actor := models.Actor{Name: "imgctx test actor"}
	if err := db.Create(&actor).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(&scene)
		db.Unscoped().Delete(&actor)
		db.Where("context IN (?)", []string{
			"imgctx-test-scene-1", "imgctx-test-unknown",
			fmt.Sprintf("act-%d", actor.ID), "icon-imgctxtest", "700x",
		}).Delete(&models.ImageProxyEntry{})
	})
	return db, scene, actor
}

func countEntries(t *testing.T, db *gorm.DB, context string) int {
	t.Helper()
	var n int
	if err := db.Model(&models.ImageProxyEntry{}).Where("context = ?", context).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestImageProxyContextRecording(t *testing.T) {
	db, scene, actor := setupImgCtxDB(t)
	upstream := newCtxTestUpstream(t)
	h := newCtxTestHandler(t, upstream, false)
	target := strings.Replace(upstream.server.URL, "://", ":/", 1) + "/cover.jpg"

	t.Run("existing scene records one row and dedupes", func(t *testing.T) {
		url := target + "?img=scene1"
		for i := 0; i < 2; i++ {
			rec := doImgRequest(h, "/img/"+scene.SceneID+"/700x/"+url)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
		}
		if n := countEntries(t, db, scene.SceneID); n != 1 {
			t.Fatalf("expected 1 row after repeat requests, got %d", n)
		}
		var e models.ImageProxyEntry
		if err := db.Where("context = ?", scene.SceneID).First(&e).Error; err != nil {
			t.Fatal(err)
		}
		if e.Options != "700x" {
			t.Fatalf("options = %q, want %q", e.Options, "700x")
		}
		wantURL := strings.Replace(url, ":/", "://", 1)
		if e.URL != wantURL {
			t.Fatalf("url = %q, want %q (incl. query string)", e.URL, wantURL)
		}
	})

	t.Run("empty options segment records empty options", func(t *testing.T) {
		rec := doImgRequest(h, "/img/"+scene.SceneID+"//"+target+"?img=noop")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var e models.ImageProxyEntry
		if err := db.Where("context = ? AND options = ''", scene.SceneID).First(&e).Error; err != nil {
			t.Fatalf("expected row with empty options: %v", err)
		}
	})

	t.Run("unknown scene context writes no row", func(t *testing.T) {
		rec := doImgRequest(h, "/img/imgctx-test-unknown/700x/"+target+"?img=unknown")
		if rec.Code != http.StatusOK {
			t.Fatalf("unknown contexts must still proxy, got %d", rec.Code)
		}
		if n := countEntries(t, db, "imgctx-test-unknown"); n != 0 {
			t.Fatalf("expected no rows, got %d", n)
		}
	})

	t.Run("zero contexts write no row", func(t *testing.T) {
		doImgRequest(h, "/img/0/700x/"+target+"?img=zero")
		doImgRequest(h, "/img/act-0/700x/"+target+"?img=zero")
		if n := countEntries(t, db, "0"); n != 0 {
			t.Fatalf("expected no rows for 0, got %d", n)
		}
		if n := countEntries(t, db, "act-0"); n != 0 {
			t.Fatalf("expected no rows for act-0, got %d", n)
		}
	})

	t.Run("actor context records", func(t *testing.T) {
		ctx := fmt.Sprintf("act-%d", actor.ID)
		rec := doImgRequest(h, "/img/"+ctx+"/700x/"+target+"?img=actor")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if n := countEntries(t, db, ctx); n != 1 {
			t.Fatalf("expected 1 row, got %d", n)
		}
	})

	t.Run("icon context records", func(t *testing.T) {
		rec := doImgRequest(h, "/img/icon-imgctxtest/20x/"+target+"?img=icon")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if n := countEntries(t, db, "icon-imgctxtest"); n != 1 {
			t.Fatalf("expected 1 row, got %d", n)
		}
	})

	t.Run("legacy-form context writes no row", func(t *testing.T) {
		// A stale /img/700x/<url> request parses as scene context "700x",
		// which does not exist — nothing may be recorded.
		doImgRequest(h, "/img/700x/"+target+"?img=legacy")
		if n := countEntries(t, db, "700x"); n != 0 {
			t.Fatalf("expected no rows for legacy context, got %d", n)
		}
	})
}

func TestImageProxyContextEmptyContext(t *testing.T) {
	upstream := newCtxTestUpstream(t)
	h := newCtxTestHandler(t, upstream, false)
	rec := doImgRequest(h, "/img//700x/http:/example.com/x.jpg")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty context, got %d", rec.Code)
	}
}

// Two different context prefixes over the same options+URL remainder must
// produce the identical upstream request and therefore share one disk cache
// entry — this is what guarantees existing cached content is reused.
func TestImageProxyContextCacheReuseAndBytePreservation(t *testing.T) {
	upstream := newCtxTestUpstream(t)
	h := newCtxTestHandler(t, upstream, true)

	remainder := "700x/" + strings.Replace(upstream.server.URL, "://", ":/", 1) + "/cover%20art.jpg?token=abc%2F123"
	for _, ctx := range []string{"0", "slr-nonexistent"} {
		rec := doImgRequest(h, "/img/"+ctx+"/"+remainder)
		if rec.Code != http.StatusOK {
			t.Fatalf("ctx %q: expected 200, got %d: %s", ctx, rec.Code, rec.Body.String())
		}
	}

	if n := atomic.LoadInt64(upstream.hits); n != 1 {
		t.Fatalf("expected 1 upstream hit (second request served from disk cache), got %d", n)
	}
	wantURI := "/cover%20art.jpg?token=abc%2F123"
	if got, _ := upstream.lastURI.Load().(string); got != wantURI {
		t.Fatalf("upstream RequestURI = %q, want %q", got, wantURI)
	}
}

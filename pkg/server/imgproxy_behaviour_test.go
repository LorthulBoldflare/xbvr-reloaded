package server

import (
	"bytes"
	"image"
	_ "image/gif"  // decode registration for DecodeConfig
	"image/jpeg" // decode registration for DecodeConfig
	_ "image/png"  // decode registration for DecodeConfig
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

// Behaviour tests for /img/<context>/<options>/<url>, one per call-site
// scenario (see plan .kilo/plans/1787781190758-image-proxy-dual-cache-tests.md).
//
// Live tests proxy real https://placehold.co/<w>x<h>[.png]?text=<text> images
// (placehold.co serves SVG by default, which imageproxy cannot transform —
// sized requests therefore use the .png form). Negative 404 tests use the
// same URL shape with the host swapped to example.com (404 for any non-root
// path; imageproxy maps upstream 404 → client 404). Live tests always run —
// they require network access by design decision.

// mockImgHandler mounts the production /img handler chain on a mux router
// exactly like StartServer does — newImageProxyBackends + newImageProxyHandler
// registered via PathPrefix("/img/"), with SkipClean(true) so remote URLs
// containing "://" survive routing — so the suite exercises the real wiring
// AND the routing layer: ForceCache, FollowRedirects, NopCache routing, the
// ForceShortCacheHandler wrapper, and PathPrefix matching. Only the disk
// cache location is swapped for a test temp dir, and skipBlocklist is true
// so loopback httptest upstreams are reachable (DenyHosts coverage lives in
// TestImageProxyBehaviourProductionBackendsDenyInternalTargets, which uses
// the production flag value).
func mockImgHandler(t *testing.T) (h http.Handler, cacheDir string) {
	t.Helper()
	cacheDir = t.TempDir()
	caching, noCache := newImageProxyBackends(cacheDir, true)
	r := mux.NewRouter()
	r.PathPrefix("/img/").Handler(newImageProxyHandler(caching, noCache))
	r.SkipClean(true)
	return r, cacheDir
}

func countCacheFiles(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func testJPEGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// Positive scenarios: every call-site pattern must proxy a real image.
// width/height assert the transformation was actually applied (0 = skip).
func TestImageProxyBehaviourLivePositive(t *testing.T) {
	h, _ := mockImgHandler(t)

	cases := []struct {
		name        string
		path        string
		contentType string // required prefix
		width       int
		height      int
	}{
		{"scene cover 700x (SceneCard/ScenePage/HereSphere/DLNA/FuckPassVR)",
			"/img/scene-beh-cover/700x/https:/placehold.co/1400x800.png?text=Scene+Cover", "image/", 700, 400},
		{"scene card 120x (QuickFind/SceneMatch/MatchSceneModal)",
			"/img/scene-beh-cover/120x/https:/placehold.co/600x400.png?text=Card", "image/", 120, 80},
		{"scene poster raw (Details.vue)",
			"/img/scene-beh-cover/raw/https:/placehold.co/600x400?text=Poster", "image/svg", 0, 0},
		{"scene preview 700,fit (ScenePlayer)",
			"/img/scene-beh-cover/700,fit/https:/placehold.co/1400x800.png?text=Preview", "image/", 700, 400},
		{"gallery thumb x40 (SceneGallery/Details)",
			"/img/scene-beh-cover/x40/https:/placehold.co/600x400.png?text=Thumb", "image/", 60, 40},
		{"gallery editor 200x",
			"/img/scene-beh-cover/200x/https:/placehold.co/600x400.png?text=Gallery", "image/", 200, 0},
		{"stashdb search result 120x (SearchStashdbScenes)",
			"/img/scene-stash-00000000-0000-0000-0000-00000000beh0/120x/https:/placehold.co/600x400.png?text=Stash", "image/", 120, 80},
		{"actor image 700x (ActorCard/ActorDetails/cast)",
			"/img/act-999999001/700x/https:/placehold.co/1400x800.png?text=Actor", "image/", 700, 400},
		{"actor search zero-id 120x (SearchStashdbActors)",
			"/img/act-0/120x/https:/placehold.co/600x400.png?text=Anon", "image/", 120, 80},
		{"site icon 20x (AltSourcesSection)",
			"/img/icon-behplace/20x/https:/placehold.co/256x256.png?text=Icon", "image/", 20, 20},
		{"scraper avatar 128x (ScrapersSection)",
			"/img/icon-behplace/128x/https:/placehold.co/256x256.png?text=Avatar", "image/", 128, 128},
		{"zero scene default",
			"/img/scene-0/700x/https:/placehold.co/1400x800.png?text=ZeroScene", "image/", 700, 400},
		{"zero icon default",
			"/img/icon-0/700x/https:/placehold.co/1400x800.png?text=ZeroIcon", "image/", 700, 400},
		{"full :// URL form (UI encodeURI output)",
			"/img/scene-beh-cover/700x/https://placehold.co/1400x800.png?text=FullForm", "image/", 700, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doImgRequest(h, tc.path)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, tc.contentType) {
				t.Fatalf("Content-Type = %q, want prefix %q", ct, tc.contentType)
			}
			// Set by ForceShortCacheHandler — proves the full production
			// handler chain (not just the bare context handler) served this.
			if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=86400" {
				t.Fatalf("Cache-Control = %q, want ForceShortCacheHandler override %q", cc, "public, max-age=86400")
			}
			if rec.Body.Len() == 0 {
				t.Fatal("empty body")
			}
			if tc.width > 0 || tc.height > 0 {
				cfg, _, err := image.DecodeConfig(rec.Body)
				if err != nil {
					t.Fatalf("proxied image does not decode: %v", err)
				}
				if tc.width > 0 && cfg.Width != tc.width {
					t.Fatalf("width = %d, want %d (transformation not applied?)", cfg.Width, tc.width)
				}
				if tc.height > 0 && cfg.Height != tc.height {
					t.Fatalf("height = %d, want %d (transformation not applied?)", cfg.Height, tc.height)
				}
			}
		})
	}
}

// Negative scenarios: upstream failures surface as non-2xx, and legacy
// pre-context URL shapes are rejected outright (they must never work again).
func TestImageProxyBehaviourLiveNegative(t *testing.T) {
	h, _ := mockImgHandler(t)

	cases := []struct {
		name string
		path string
		want int
	}{
		{"upstream 404 via example.com, sized",
			"/img/scene-beh-cover/700x/https:/example.com/700x400.png?text=Gone", http.StatusNotFound},
		{"upstream 404 via example.com, raw",
			"/img/scene-beh-cover/raw/https:/example.com/600x400?text=Gone", http.StatusNotFound},
		{"legacy options-only form",
			"/img/700x/placehold.co/600x400.png?text=Legacy", http.StatusBadRequest},
		{"legacy options-only form with scheme",
			"/img/700x/https:/placehold.co/600x400.png?text=Legacy", http.StatusBadRequest},
		{"legacy 0x0 options form",
			"/img/0x0/https:/placehold.co/600x400.png?text=Legacy", http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doImgRequest(h, tc.path)
			if rec.Code != tc.want {
				t.Fatalf("expected %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}


// The mock router must behave like StartServer's mux: only the /img/ subtree
// is served; everything else (including bare "/img", which PathPrefix("/img/")
// does not match) 404s at the routing layer.
func TestImageProxyMockRouterOnlyServesImgSubtree(t *testing.T) {
	h, _ := mockImgHandler(t)

	for _, path := range []string{"/img", "/other/x", "/imghm/x"} {
		rec := doImgRequest(h, path)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("path %q: expected router-level 404, got %d", path, rec.Code)
		}
	}
}

// Attributed requests go to the caching backend and land in the disk cache.
func TestImageProxyBehaviourLiveAttributedPopulatesDiskCache(t *testing.T) {
	h, dir := mockImgHandler(t)

	rec := doImgRequest(h, "/img/scene-beh-cache/700x/https:/placehold.co/600x400.png?text=CacheMe")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if n := countCacheFiles(t, dir); n == 0 {
		t.Fatal("expected at least one disk cache entry after attributed request")
	}
}

// Zero-id requests go to the NopCache backend and never touch the disk cache.
func TestImageProxyBehaviourLiveZeroIdLeavesDiskCacheEmpty(t *testing.T) {
	h, dir := mockImgHandler(t)

	for _, ctx := range []string{"scene-0", "act-0", "icon-0"} {
		rec := doImgRequest(h, "/img/"+ctx+"/700x/https:/placehold.co/600x400.png?text=NoCache+"+ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("ctx %q: expected 200, got %d: %s", ctx, rec.Code, rec.Body.String())
		}
	}
	if n := countCacheFiles(t, dir); n != 0 {
		t.Fatalf("expected empty disk cache after zero-id-only requests, found %d entries", n)
	}
}

// imageproxy cannot transform SVG: sized options on an SVG upstream do not
// fail — the original bytes pass through untransformed (Transform errors are
// logged and ignored). This is why sized live tests must use placehold.co's
// .png form. raw passes it through as well.
func TestImageProxySizedOptionsOnSVGPassThroughUntransformed(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(svg))
	}))
	defer upstream.Close()

	h := newCtxTestHandler(t, true)
	target := strings.Replace(upstream.URL, "://", ":/", 1) + "/icon.svg"

	for _, options := range []string{"700x", "raw"} {
		t.Run(options, func(t *testing.T) {
			rec := doImgRequest(h, "/img/scene-svg-test/"+options+"/"+target+"?v="+options)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/svg") {
				t.Fatalf("Content-Type = %q, want image/svg*", ct)
			}
			if !strings.Contains(rec.Body.String(), "<svg") {
				t.Fatal("expected original SVG body to pass through untransformed")
			}
		})
	}
}

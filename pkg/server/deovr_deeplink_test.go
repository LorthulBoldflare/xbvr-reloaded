package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful/v3"

	"github.com/xbapps/xbvr/pkg/api"
	"github.com/xbapps/xbvr/pkg/config"
)

func dispatchDeoVRDeeplink(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	container := restful.NewContainer()
	container.Filter(apiAuthFilter)
	// Register the real resource so the filter's path grammar is tested
	// against the production route, not a stub copy of it. The handler runs
	// without a seeded DB, so accepted requests surface as 404 (unknown
	// scene) — the filter decision is what these tests assert on.
	container.Add(api.DeoVRDeeplinkResource{}.WebService())

	rec := httptest.NewRecorder()
	container.Dispatch(rec, req)
	return rec
}

// The per-scene deeplink token (?token=) authenticates /api/deovr/<id>.json
// for exactly that scene, because DeoVR follows deep links with a plain GET
// and can present neither Basic credentials nor the player-session cookie.
func TestAPIAuthFilterDeoVRDeeplinkToken(t *testing.T) {
	setupPlayerSessionTest(t)

	token123 := config.DeoVRDeeplinkToken(123)
	if token123 == "" {
		t.Fatal("expected deeplink token while player auth enabled")
	}

	deeplinkReq := func(path string) *http.Request {
		return httptest.NewRequest(http.MethodGet, "http://xbvr.local"+path, nil)
	}

	t.Run("valid per-scene token accepted", func(t *testing.T) {
		rec := dispatchDeoVRDeeplink(t, deeplinkReq("/api/deovr/123.json?token="+token123))
		// Passed the filter; the real handler 404s without a seeded DB.
		if rec.Code == http.StatusUnauthorized {
			t.Fatal("valid deeplink token was rejected with 401")
		}
	})

	t.Run("token for another scene rejected", func(t *testing.T) {
		rec := dispatchDeoVRDeeplink(t, deeplinkReq("/api/deovr/456.json?token="+token123))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 with cross-scene token, got %d", rec.Code)
		}
	})

	t.Run("wrong token rejected", func(t *testing.T) {
		rec := dispatchDeoVRDeeplink(t, deeplinkReq("/api/deovr/123.json?token=deadbeef"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 with wrong token, got %d", rec.Code)
		}
	})

	t.Run("missing token rejected", func(t *testing.T) {
		rec := dispatchDeoVRDeeplink(t, deeplinkReq("/api/deovr/123.json"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without token, got %d", rec.Code)
		}
	})

	t.Run("malformed scene segment falls through to normal auth", func(t *testing.T) {
		rec := dispatchDeoVRDeeplink(t, deeplinkReq("/api/deovr/abc.json?token="+token123))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for malformed scene id, got %d", rec.Code)
		}
	})

	t.Run("token not accepted on other api paths", func(t *testing.T) {
		rec := dispatchAPI(t, deeplinkReq("/api/ping?token="+token123))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for token on /api/ping, got %d", rec.Code)
		}
	})

	t.Run("token ignored when player auth disabled", func(t *testing.T) {
		config.Config.Interfaces.DeoVR.AuthEnabled = false
		defer func() { config.Config.Interfaces.DeoVR.AuthEnabled = true }()
		rec := dispatchDeoVRDeeplink(t, deeplinkReq("/api/deovr/123.json?token="+token123))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 while player auth disabled, got %d", rec.Code)
		}
	})

	t.Run("owner session cookie still works alongside token", func(t *testing.T) {
		req := withPlayerCookie(deeplinkReq("/api/deovr/123.json"), config.PlayerSessionToken())
		rec := dispatchDeoVRDeeplink(t, req)
		if rec.Code == http.StatusUnauthorized {
			t.Fatal("player session cookie was rejected with 401")
		}
	})
}

func TestDeoVRDeeplinkSceneIDParsing(t *testing.T) {
	tests := []struct {
		path    string
		wantID  uint
		wantOK  bool
	}{
		{"/api/deovr/123.json", 123, true},
		{"/api/deovr/1.json", 1, true},
		{"/api/deovr/123", 0, false},          // missing .json suffix
		{"/api/deovr/abc.json", 0, false},     // non-numeric id
		{"/api/deovr/123.json.json", 0, false},// non-numeric after suffix strip
		{"/api/deovr/", 0, false},
		{"/api/scene/123.json", 0, false},     // wrong prefix
		{"/deovr/123.json", 0, false},         // player endpoint, not deeplink
	}
	for _, tt := range tests {
		gotID, gotOK := deoVRDeeplinkSceneID(tt.path)
		if gotID != tt.wantID || gotOK != tt.wantOK {
			t.Errorf("deoVRDeeplinkSceneID(%q) = (%d, %v), want (%d, %v)", tt.path, gotID, gotOK, tt.wantID, tt.wantOK)
		}
	}
}

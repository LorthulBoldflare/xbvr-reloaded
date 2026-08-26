package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/crypto/bcrypt"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
)

// TestMain keeps authlog output from tests out of the real
// $HOME/xbvr_auth.log.
func TestMain(m *testing.M) {
	os.Setenv("XBVR_AUTH_LOG", filepath.Join(os.TempDir(), "xbvr_auth_server_test.log"))
	os.Exit(m.Run())
}

// setupPlayerSessionTest enables UI auth and player auth with a stable
// token, restoring both afterwards.
func setupPlayerSessionTest(t *testing.T) {
	t.Helper()

	origUser, origPass, origNoAuth := common.EnvConfig.UIUsername, common.EnvConfig.UIPassword, common.EnvConfig.NoAPIAuth
	origDeo := config.Config.Interfaces.DeoVR
	t.Cleanup(func() {
		common.EnvConfig.UIUsername, common.EnvConfig.UIPassword, common.EnvConfig.NoAPIAuth = origUser, origPass, origNoAuth
		config.Config.Interfaces.DeoVR = origDeo
	})

	common.EnvConfig.UIUsername = "admin"
	common.EnvConfig.UIPassword = "secret"
	common.EnvConfig.NoAPIAuth = false

	config.Config.Interfaces.DeoVR.AuthEnabled = true
	config.Config.Interfaces.DeoVR.Username = "player"
	config.Config.Interfaces.DeoVR.Password = "$2a$10$somebcrypthashvalue" // token key; never verified here
}

func dispatchAPI(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	container := restful.NewContainer()
	container.Filter(apiAuthFilter)
	ws := new(restful.WebService)
	ws.Route(ws.GET("/api/ping").To(func(req *restful.Request, resp *restful.Response) {
		resp.WriteHeader(http.StatusOK)
	}))
	ws.Route(ws.POST("/api/ping").To(func(req *restful.Request, resp *restful.Response) {
		resp.WriteHeader(http.StatusOK)
	}))
	container.Add(ws)

	rec := httptest.NewRecorder()
	container.Dispatch(rec, req)
	return rec
}

func withPlayerCookie(req *http.Request, value string) *http.Request {
	req.AddCookie(&http.Cookie{Name: config.PlayerSessionCookieName, Value: value})
	return req
}

func TestHasValidPlayerSession(t *testing.T) {
	setupPlayerSessionTest(t)

	t.Run("valid token accepted", func(t *testing.T) {
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil), config.PlayerSessionToken())
		if !hasValidPlayerSession(req) {
			t.Fatal("expected valid player session cookie to be accepted")
		}
	})

	t.Run("wrong token rejected", func(t *testing.T) {
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil), "deadbeef")
		if hasValidPlayerSession(req) {
			t.Fatal("expected wrong cookie value to be rejected")
		}
	})

	t.Run("no cookie rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil)
		if hasValidPlayerSession(req) {
			t.Fatal("expected missing cookie to be rejected")
		}
	})

	t.Run("rejected when player auth disabled", func(t *testing.T) {
		config.Config.Interfaces.DeoVR.AuthEnabled = false
		defer func() { config.Config.Interfaces.DeoVR.AuthEnabled = true }()
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil), config.PlayerSessionToken())
		if hasValidPlayerSession(req) {
			t.Fatal("expected cookie to be rejected while player auth disabled")
		}
	})
}

func TestAPIAuthFilterPlayerCookie(t *testing.T) {
	setupPlayerSessionTest(t)

	t.Run("valid cookie passes without basic auth", func(t *testing.T) {
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil), config.PlayerSessionToken())
		rec := dispatchAPI(t, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 with valid cookie, got %d", rec.Code)
		}
	})

	t.Run("invalid cookie falls back to 401", func(t *testing.T) {
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil), "deadbeef")
		rec := dispatchAPI(t, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 with invalid cookie, got %d", rec.Code)
		}
	})

	t.Run("cookie does not unlock exempt dms path check", func(t *testing.T) {
		// /api/dms/* stays exempt regardless of cookies
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/dms/file/1", nil), "deadbeef")
		rec := dispatchAPI(t, req)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected /api/dms/* to bypass auth, got 401")
		}
	})
}

func TestAPIAuthFilterPlayerBasicAuth(t *testing.T) {
	setupPlayerSessionTest(t)

	// Real bcrypt hash so the password comparison can succeed.
	hash, err := bcrypt.GenerateFromPassword([]byte("player-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	config.Config.Interfaces.DeoVR.Password = string(hash)

	t.Run("player credentials accepted and mint cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil)
		req.SetBasicAuth("player", "player-pass")
		rec := dispatchAPI(t, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 with player credentials, got %d", rec.Code)
		}
		var c *http.Cookie
		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == config.PlayerSessionCookieName {
				c = cookie
			}
		}
		if c == nil {
			t.Fatal("expected session cookie to be minted on player basic auth")
		}
		if c.Value != config.PlayerSessionToken() {
			t.Fatal("cookie value does not match PlayerSessionToken()")
		}
	})

	t.Run("wrong player password rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil)
		req.SetBasicAuth("player", "wrong")
		rec := dispatchAPI(t, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("UI credentials still accepted, no cookie minted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil)
		req.SetBasicAuth("admin", "secret")
		rec := dispatchAPI(t, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 with UI credentials, got %d", rec.Code)
		}
		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == config.PlayerSessionCookieName {
				t.Fatal("session cookie must not be minted for UI credentials")
			}
		}
	})

	t.Run("player basic rejected when player auth disabled", func(t *testing.T) {
		config.Config.Interfaces.DeoVR.AuthEnabled = false
		defer func() { config.Config.Interfaces.DeoVR.AuthEnabled = true }()
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil)
		req.SetBasicAuth("player", "player-pass")
		rec := dispatchAPI(t, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 when player auth disabled, got %d", rec.Code)
		}
	})
}

func TestAuthDataSummary(t *testing.T) {
	setupPlayerSessionTest(t)

	t.Run("no auth data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/dms/file/1", nil)
		if got := authDataSummary(req); got != "none" {
			t.Fatalf("authDataSummary() = %q, want %q", got, "none")
		}
	})

	t.Run("valid player cookie", func(t *testing.T) {
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/dms/file/1", nil), config.PlayerSessionToken())
		if got := authDataSummary(req); got != "player-cookie valid" {
			t.Fatalf("authDataSummary() = %q, want %q", got, "player-cookie valid")
		}
	})

	t.Run("invalid player cookie", func(t *testing.T) {
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/dms/file/1", nil), "deadbeef")
		if got := authDataSummary(req); got != "player-cookie INVALID" {
			t.Fatalf("authDataSummary() = %q, want %q", got, "player-cookie INVALID")
		}
	})

	t.Run("basic auth header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/dms/file/1", nil)
		req.SetBasicAuth("someone", "whatever")
		want := `basic user="someone"`
		if got := authDataSummary(req); got != want {
			t.Fatalf("authDataSummary() = %q, want %q", got, want)
		}
	})
}

// Origin mismatches are logged but never block the request: players are
// non-standard browsers whose Origin headers cannot be relied upon.
func TestAPIAuthFilterOriginMismatch(t *testing.T) {
	setupPlayerSessionTest(t)

	postWithOrigin := func(origin string) *http.Request {
		req := withPlayerCookie(httptest.NewRequest(http.MethodPost, "http://xbvr.local/api/ping", nil), config.PlayerSessionToken())
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		return req
	}

	t.Run("cross-origin mutating request allowed", func(t *testing.T) {
		rec := dispatchAPI(t, postWithOrigin("http://evil.example.com"))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for cross-origin POST (warn-and-allow), got %d", rec.Code)
		}
	})

	t.Run("same-origin mutating request passes", func(t *testing.T) {
		rec := dispatchAPI(t, postWithOrigin("http://xbvr.local"))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for same-origin POST, got %d", rec.Code)
		}
	})

	t.Run("mutating request without origin passes", func(t *testing.T) {
		rec := dispatchAPI(t, postWithOrigin(""))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for POST without Origin, got %d", rec.Code)
		}
	})

	t.Run("cross-origin read-only request passes", func(t *testing.T) {
		req := withPlayerCookie(httptest.NewRequest(http.MethodGet, "http://xbvr.local/api/ping", nil), config.PlayerSessionToken())
		req.Header.Set("Origin", "http://evil.example.com")
		rec := dispatchAPI(t, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for cross-origin GET, got %d", rec.Code)
		}
	})
}

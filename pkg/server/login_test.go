package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/xbapps/xbvr/pkg/config"
)

func TestIsPlayerClient(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		rw   string
		want bool
	}{
		{"DeoVR API stack", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/40.0.2214.111 Chrome/40.0.2214.111 Safari/537.36 HMD/12 [DEO15.8.3888]/meta-store", "", true},
		{"DeoVR playback (AVPro/ExoPlayer)", "AVProMobileVideo/15.8.3888 (Linux;Android 14) ExoPlayerLib/1.4.1", "", true},
		{"HereSphere API stack", "HereSphere/0.8 Quest2", "", true},
		{"HereSphere media stack (Cronet)", "com.heresphere.vrvideoplayer/29 (Linux; U; Android 14; en_US; Quest 2; Build/UP1A.231005.007.A1; Cronet/95.0.4638.50)", "", true},
		{"HereSphere browser via X-Requested-With", "Mozilla/5.0 (X11; Linux x86_64; Android 12; Quest 3) AppleWebKit/537.36", "com.heresphere.vrvideoplayer", true},
		{"DVRHyper", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/605.1.15 (KHTML, like Gecko) Ubuntu Chromium/147.0.0.0 Chrome/147.0.0.0 Safari/605.1.15 HMD/12 [HYPER1.12.2092]/meta-store", "", true},
		{"DVRHyper via X-Requested-With", "Mozilla/5.0", "com.DeoVR.Hyper", true},
		{"stagefright alone is not enough", "stagefright/1.2 (Linux;Android 14)", "", false},
		{"desktop Chrome", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36", "", false},
		{"curl", "curl/8.7.1", "", false},
		{"empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/", nil)
			if tt.ua != "" {
				req.Header.Set("User-Agent", tt.ua)
			}
			if tt.rw != "" {
				req.Header.Set("X-Requested-With", tt.rw)
			}
			if got := isPlayerClient(req); got != tt.want {
				t.Fatalf("isPlayerClient(ua=%q rw=%q) = %v, want %v", tt.ua, tt.rw, got, tt.want)
			}
		})
	}
}

func setupLoginTest(t *testing.T) {
	t.Helper()
	orig := config.Config.Interfaces.DeoVR
	t.Cleanup(func() { config.Config.Interfaces.DeoVR = orig })

	hash, err := bcrypt.GenerateFromPassword([]byte("player-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	config.Config.Interfaces.DeoVR.AuthEnabled = true
	config.Config.Interfaces.DeoVR.Username = "player"
	config.Config.Interfaces.DeoVR.Password = string(hash)
}

const deoUA = "Mozilla/5.0 HMD/12 [DEO15.8.3888]/meta-store"

func TestLoginHandler(t *testing.T) {
	setupLoginTest(t)

	t.Run("GET as player serves the form", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/login", nil)
		req.Header.Set("User-Agent", deoUA)
		rec := httptest.NewRecorder()
		loginHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `name="username"`) || !strings.Contains(body, `name="password"`) {
			t.Fatalf("login page must contain username and password fields:\n%s", body)
		}
		if strings.Contains(strings.ToLower(body), "xbvr") {
			t.Fatal("login page must be non-descript (no xbvr branding)")
		}
	})

	t.Run("GET as non-player is forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/login", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/151.0.0.0")
		rec := httptest.NewRecorder()
		loginHandler(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("POST with valid player credentials sets cookie and redirects", func(t *testing.T) {
		form := url.Values{"username": {"player"}, "password": {"player-pass"}}
		req := httptest.NewRequest(http.MethodPost, "http://xbvr.local/login", strings.NewReader(form.Encode()))
		req.Header.Set("User-Agent", deoUA)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		loginHandler(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/ui/" {
			t.Fatalf("expected redirect to /ui/, got %q", loc)
		}
		var c *http.Cookie
		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == config.PlayerSessionCookieName {
				c = cookie
			}
		}
		if c == nil {
			t.Fatal("expected xbvr_player_session cookie to be set")
		}
		if c.Value != config.PlayerSessionToken() {
			t.Fatal("cookie value does not match PlayerSessionToken()")
		}
	})

	t.Run("POST with wrong password re-serves form with 401", func(t *testing.T) {
		form := url.Values{"username": {"player"}, "password": {"wrong"}}
		req := httptest.NewRequest(http.MethodPost, "http://xbvr.local/login", strings.NewReader(form.Encode()))
		req.Header.Set("User-Agent", deoUA)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		loginHandler(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == config.PlayerSessionCookieName {
				t.Fatal("no cookie may be set on failed login")
			}
		}
	})

	t.Run("POST as non-player is forbidden even with valid credentials", func(t *testing.T) {
		form := url.Values{"username": {"player"}, "password": {"player-pass"}}
		req := httptest.NewRequest(http.MethodPost, "http://xbvr.local/login", strings.NewReader(form.Encode()))
		req.Header.Set("User-Agent", "curl/8.7.1")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		loginHandler(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("GET with valid session cookie redirects straight to /ui/", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://xbvr.local/login", nil)
		req.Header.Set("User-Agent", deoUA)
		req.AddCookie(&http.Cookie{Name: config.PlayerSessionCookieName, Value: config.PlayerSessionToken()})
		rec := httptest.NewRecorder()
		loginHandler(rec, req)
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/ui/" {
			t.Fatalf("expected 303 to /ui/, got %d %q", rec.Code, rec.Header().Get("Location"))
		}
	})
}

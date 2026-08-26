package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/crypto/bcrypt"

	"github.com/xbapps/xbvr/pkg/config"
)

const playerTestPassword = "player-pass"

func setupPlayerAuth(t *testing.T, enabled bool) {
	t.Helper()
	orig := config.Config.Interfaces.DeoVR
	t.Cleanup(func() { config.Config.Interfaces.DeoVR = orig })

	hash, err := bcrypt.GenerateFromPassword([]byte(playerTestPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	config.Config.Interfaces.DeoVR.AuthEnabled = enabled
	config.Config.Interfaces.DeoVR.Username = "player"
	config.Config.Interfaces.DeoVR.Password = string(hash)
}

func dispatchWithFilter(t *testing.T, path string, filter restful.FilterFunction, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	container := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Path(path)
	ws.Route(ws.POST("/").Filter(filter).To(func(req *restful.Request, resp *restful.Response) {
		resp.WriteHeader(http.StatusOK)
	}))
	container.Add(ws)

	rec := httptest.NewRecorder()
	container.Dispatch(rec, req)
	return rec
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == config.PlayerSessionCookieName {
			return c
		}
	}
	return nil
}

func TestDeoVRFilterMintsSessionCookie(t *testing.T) {
	setupPlayerAuth(t, true)

	deoReq := func(body string) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/deovr/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	t.Run("successful login sets cookie", func(t *testing.T) {
		rec := dispatchWithFilter(t, "/deovr", restfulAuthFilter, deoReq("login=player&password="+playerTestPassword))
		c := sessionCookie(t, rec)
		if c == nil {
			t.Fatal("expected session cookie on successful DeoVR auth")
		}
		if c.Value != config.PlayerSessionToken() {
			t.Fatal("cookie value does not match PlayerSessionToken()")
		}
		if c.Path != "/" || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
			t.Fatalf("unexpected cookie attributes: %+v", c)
		}
	})

	t.Run("failed login sets no cookie", func(t *testing.T) {
		rec := dispatchWithFilter(t, "/deovr", restfulAuthFilter, deoReq("login=player&password=wrong"))
		if c := sessionCookie(t, rec); c != nil {
			t.Fatalf("unexpected cookie on failed auth: %+v", c)
		}
	})

	t.Run("missing credentials set no cookie", func(t *testing.T) {
		rec := dispatchWithFilter(t, "/deovr", restfulAuthFilter, deoReq(""))
		if c := sessionCookie(t, rec); c != nil {
			t.Fatalf("unexpected cookie without credentials: %+v", c)
		}
	})
}

func TestHeresphereFilterMintsSessionCookie(t *testing.T) {
	setupPlayerAuth(t, true)

	hsReq := func(body string) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/heresphere/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	t.Run("successful login sets cookie", func(t *testing.T) {
		rec := dispatchWithFilter(t, "/heresphere", HeresphereAuthFilter,
			hsReq(`{"username":"player","password":"`+playerTestPassword+`"}`))
		c := sessionCookie(t, rec)
		if c == nil {
			t.Fatal("expected session cookie on successful HereSphere auth")
		}
		if c.Value != config.PlayerSessionToken() {
			t.Fatal("cookie value does not match PlayerSessionToken()")
		}
	})

	t.Run("failed login sets no cookie", func(t *testing.T) {
		rec := dispatchWithFilter(t, "/heresphere", HeresphereAuthFilter,
			hsReq(`{"username":"player","password":"wrong"}`))
		if c := sessionCookie(t, rec); c != nil {
			t.Fatalf("unexpected cookie on failed auth: %+v", c)
		}
	})
}

func TestNoSessionCookieWhenPlayerAuthDisabled(t *testing.T) {
	setupPlayerAuth(t, false)

	req := httptest.NewRequest(http.MethodPost, "/deovr/", strings.NewReader("login=player&password="+playerTestPassword))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := dispatchWithFilter(t, "/deovr", restfulAuthFilter, req)
	if c := sessionCookie(t, rec); c != nil {
		t.Fatalf("unexpected cookie while player auth disabled: %+v", c)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected request to pass through when player auth disabled, got %d", rec.Code)
	}
}

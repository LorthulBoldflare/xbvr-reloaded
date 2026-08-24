package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful/v3"

	"github.com/xbapps/xbvr/pkg/common"
)

func runFilter(t *testing.T, path string, withCreds bool) *httptest.ResponseRecorder {
	t.Helper()

	container := restful.NewContainer()
	container.Filter(apiAuthFilter)
	ws := new(restful.WebService)
	ws.Route(ws.GET("/x").To(func(req *restful.Request, resp *restful.Response) {
		resp.WriteHeader(http.StatusOK)
	}))
	container.Add(ws)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if withCreds {
		req.SetBasicAuth(common.EnvConfig.UIUsername, common.EnvConfig.UIPassword)
	}
	rec := httptest.NewRecorder()
	container.Dispatch(rec, req)
	return rec
}

func TestAPIAuthFilter(t *testing.T) {
	origUser, origPass, origNoAuth := common.EnvConfig.UIUsername, common.EnvConfig.UIPassword, common.EnvConfig.NoAPIAuth
	defer func() {
		common.EnvConfig.UIUsername, common.EnvConfig.UIPassword, common.EnvConfig.NoAPIAuth = origUser, origPass, origNoAuth
	}()

	common.EnvConfig.UIUsername = "admin"
	common.EnvConfig.UIPassword = "secret"
	common.EnvConfig.NoAPIAuth = false

	// ws root path doesn't matter for routing here; use full paths
	t.Run("rejects unauthenticated /api request", func(t *testing.T) {
		rec := runFilter(t, "/api/options/state", false)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		if rec.Header().Get("WWW-Authenticate") == "" {
			t.Fatal("expected WWW-Authenticate challenge header")
		}
	})

	t.Run("accepts valid credentials", func(t *testing.T) {
		rec := runFilter(t, "/api/options/state", true)
		// 404 from the mux is fine — auth passed, route simply not registered
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected auth to pass, got 401")
		}
	})

	t.Run("player endpoints unaffected", func(t *testing.T) {
		rec := runFilter(t, "/deovr", false)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected /deovr to bypass API auth, got 401")
		}
		rec = runFilter(t, "/heresphere", false)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected /heresphere to bypass API auth, got 401")
		}
	})

	t.Run("player media endpoints unaffected", func(t *testing.T) {
		// DeoVR/HereSphere stream video, funscripts and previews from
		// /api/dms/* and cannot send Authorization headers on those requests
		for _, path := range []string{"/api/dms/file/123", "/api/dms/file/123/script.funscript", "/api/dms/preview/45", "/api/dms/heatmap/67"} {
			rec := runFilter(t, path, false)
			if rec.Code == http.StatusUnauthorized {
				t.Fatalf("expected %s to bypass API auth, got 401", path)
			}
		}
	})

	t.Run("openapi spec unaffected", func(t *testing.T) {
		rec := runFilter(t, "/api.json", false)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected /api.json to bypass API auth, got 401")
		}
	})

	t.Run("disabled when UI auth is off", func(t *testing.T) {
		common.EnvConfig.UIUsername = ""
		common.EnvConfig.UIPassword = ""
		defer func() { common.EnvConfig.UIUsername, common.EnvConfig.UIPassword = "admin", "secret" }()
		rec := runFilter(t, "/api/options/state", false)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected no auth when UI auth disabled, got 401")
		}
	})

	t.Run("disabled via XBVR_NO_API_AUTH", func(t *testing.T) {
		common.EnvConfig.NoAPIAuth = true
		defer func() { common.EnvConfig.NoAPIAuth = false }()
		rec := runFilter(t, "/api/options/state", false)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected no auth with NoAPIAuth, got 401")
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/options/state", nil)
		req.SetBasicAuth("admin", "wrong")
		rec := httptest.NewRecorder()
		container := restful.NewContainer()
		container.Filter(apiAuthFilter)
		container.Dispatch(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}

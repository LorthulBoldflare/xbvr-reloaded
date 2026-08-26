package server

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"sync"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/crypto/bcrypt"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
)

// hasValidPlayerSession reports whether the request carries a valid
// player-session cookie — the stable credential-derived token minted by the
// /deovr and /heresphere auth filters (see config.PlayerSessionToken). It is
// accepted as an alternative to HTTP Basic Auth on the Web UI surfaces, so a
// native player login also unlocks the Web UI in the same browser. Always
// false when player auth is disabled.
func hasValidPlayerSession(r *http.Request) bool {
	token := config.PlayerSessionToken()
	if token == "" {
		return false
	}
	c, err := r.Cookie(config.PlayerSessionCookieName)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.Value), []byte(token)) == 1
}

// apiAuthFilter enforces HTTP Basic Auth on all /api/* routes whenever UI
// auth is enabled (UI_USERNAME/UI_PASSWORD set). A valid player-session
// cookie (minted by a native DeoVR/HereSphere login) is accepted as an
// alternative credential. Player endpoints (/deovr, /heresphere) are rooted
// outside /api and keep their own auth toggles.
// /api/dms/* is exempt: it is the media surface the players stream from
// (video, funscripts, previews), and neither DeoVR nor HereSphere can send
// Authorization headers on those requests.
// Set XBVR_NO_API_AUTH=1 to disable API-scoped authentication.
func apiAuthFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if !common.IsUIAuthEnabled() || common.EnvConfig.NoAPIAuth ||
		!strings.HasPrefix(req.Request.URL.Path, "/api/") ||
		strings.HasPrefix(req.Request.URL.Path, "/api/dms/") {
		chain.ProcessFilter(req, resp)
		return
	}

	// CSRF note for the cookie path: unlike Basic Auth credentials, cookies
	// are attached by the browser to cross-site requests. Players are
	// non-standard browsers whose Origin headers cannot be relied upon, so
	// a mismatch is logged but the request is allowed. SameSite=Lax on the
	// cookie remains the effective layer in standard browsers.
	if isMutatingMethod(req.Request.Method) && !browserOriginMatchesHost(req.Request) {
		log.Warnf("mutating API request %s %s with non-matching Origin %q (Host %q) — allowing (non-standard browser)",
			req.Request.Method, req.Request.URL.Path, req.Request.Header.Get("Origin"), req.Request.Host)
	}

	if hasValidPlayerSession(req.Request) {
		chain.ProcessFilter(req, resp)
		return
	}

	user, password, ok := req.Request.BasicAuth()
	if !ok || user != common.EnvConfig.UIUsername || !checkUIPassword(password) {
		resp.AddHeader("WWW-Authenticate", `Basic realm="default"`)
		resp.WriteErrorString(http.StatusUnauthorized, "401: Unauthorized")
		return
	}

	chain.ProcessFilter(req, resp)
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	return true
}

// checkUIPassword compares against a cached bcrypt hash of the configured
// password — hashing on every request (previous behavior) costs ~60-250ms of
// CPU per API call. The cache is keyed on the configured password so config
// changes (and tests) take effect immediately.
var cachedUIAuth struct {
	sync.Mutex
	password string
	hash     []byte
}

func checkUIPassword(password string) bool {
	cachedUIAuth.Lock()
	if cachedUIAuth.password != common.EnvConfig.UIPassword {
		hash, err := bcrypt.GenerateFromPassword([]byte(common.EnvConfig.UIPassword), bcrypt.DefaultCost)
		if err != nil {
			cachedUIAuth.Unlock()
			return false
		}
		cachedUIAuth.password = common.EnvConfig.UIPassword
		cachedUIAuth.hash = hash
	}
	hash := cachedUIAuth.hash
	cachedUIAuth.Unlock()

	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}

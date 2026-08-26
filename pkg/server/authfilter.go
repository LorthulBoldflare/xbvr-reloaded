package server

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/crypto/bcrypt"

	"github.com/xbapps/xbvr/pkg/authlog"
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
	if !strings.HasPrefix(req.Request.URL.Path, "/api/") {
		chain.ProcessFilter(req, resp)
		return
	}

	if strings.HasPrefix(req.Request.URL.Path, "/api/dms/") {
		// Exempt media surface. Bodies are binary streams, but log the full
		// request headers plus any authentication data the player happens to
		// send — this tells us whether cookie auth could be required on this
		// surface in the future.
		authlog.Request("dms", req.Request, nil)
		authlog.Event("dms", "auth data on exempt path: %s", authDataSummary(req.Request))
		chain.ProcessFilter(req, resp)
		return
	}

	if !common.IsUIAuthEnabled() || common.EnvConfig.NoAPIAuth {
		authlog.Request("api", req.Request, nil)
		authlog.Event("api", "auth data on exempt path (auth disabled): %s", authDataSummary(req.Request))
		chain.ProcessFilter(req, resp)
		return
	}

	// Buffer the raw body for the auth log and restore it for downstream
	// consumers.
	var rawBody []byte
	if req.Request.Body != nil && req.Request.Body != http.NoBody {
		rawBody, _ = io.ReadAll(req.Request.Body)
		req.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
	}
	authlog.Request("api", req.Request, rawBody)

	// CSRF note for the cookie path: unlike Basic Auth credentials, cookies
	// are attached by the browser to cross-site requests. Players are
	// non-standard browsers whose Origin headers cannot be relied upon, so
	// a mismatch is logged but the request is allowed. SameSite=Lax on the
	// cookie remains the effective layer in standard browsers.
	if isMutatingMethod(req.Request.Method) && !browserOriginMatchesHost(req.Request) {
		log.Warnf("mutating API request %s %s with non-matching Origin %q (Host %q) — allowing (non-standard browser)",
			req.Request.Method, req.Request.URL.Path, req.Request.Header.Get("Origin"), req.Request.Host)
		authlog.Event("api", "origin mismatch on mutating request (allowed): Origin %q vs Host %q",
			req.Request.Header.Get("Origin"), req.Request.Host)
	}

	if hasValidPlayerSession(req.Request) {
		authlog.Event("api", "auth method=cookie result=accepted")
		chain.ProcessFilter(req, resp)
		return
	}
	if _, err := req.Request.Cookie(config.PlayerSessionCookieName); err == nil {
		authlog.Event("api", "auth method=cookie result=rejected (cookie present but invalid)")
	}

	user, password, ok := req.Request.BasicAuth()
	basicUI := ok && user == common.EnvConfig.UIUsername && checkUIPassword(password)
	basicPlayer := !basicUI && checkPlayerBasicAuth(user, password)
	switch {
	case !ok:
		authlog.Event("api", "auth result=denied (no credentials)")
	case basicUI:
		authlog.Event("api", "auth method=basic user=%q result=success", user)
	case basicPlayer:
		authlog.Event("api", "auth method=basic user=%q result=success (player credentials)", user)
	default:
		authlog.Event("api", "auth method=basic user=%q result=failed", user)
	}
	if !basicUI && !basicPlayer {
		resp.AddHeader("WWW-Authenticate", `Basic realm="default"`)
		resp.WriteErrorString(http.StatusUnauthorized, "401: Unauthorized")
		return
	}
	if basicPlayer {
		// Mint the player-session cookie so subsequent requests from this
		// browser authenticate without the Basic prompt.
		setPlayerSessionCookie("api", resp.ResponseWriter)
	}

	chain.ProcessFilter(req, resp)
}

// checkPlayerBasicAuth reports whether the given Basic credentials match the
// configured DeoVR/HereSphere player credential pair (verified against the
// stored bcrypt hash). Always false when player auth is disabled.
func checkPlayerBasicAuth(user, password string) bool {
	if user == "" || !config.PlayerAuthEnabled() || user != config.Config.Interfaces.DeoVR.Username {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(config.Config.Interfaces.DeoVR.Password), []byte(password)) == nil
}

// setPlayerSessionCookie mints the player-session cookie on a plain
// http.ResponseWriter (Web UI surfaces) and logs it.
func setPlayerSessionCookie(component string, w http.ResponseWriter) {
	cookie := config.PlayerSessionCookie()
	if cookie == nil {
		return
	}
	http.SetCookie(w, cookie)
	authlog.Event(component, "minted session cookie %s=%s", config.PlayerSessionCookieName, cookie.Value)
}

// authDataSummary describes any authentication material present on a
// request hitting a path where auth is not (currently) required — used to
// learn whether players could authenticate there if it ever became required.
func authDataSummary(r *http.Request) string {
	var parts []string
	if user, _, ok := r.BasicAuth(); ok {
		parts = append(parts, fmt.Sprintf("basic user=%q", user))
	}
	if _, err := r.Cookie(config.PlayerSessionCookieName); err == nil {
		if hasValidPlayerSession(r) {
			parts = append(parts, "player-cookie valid")
		} else {
			parts = append(parts, "player-cookie INVALID")
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
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

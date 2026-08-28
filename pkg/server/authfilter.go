package server

import (
	"bytes"
	"crypto/subtle"
	"io"
	"net/http"
	"strconv"
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
		// Exempt media surface. Bodies are binary streams, but the entry
		// still records any authentication data the player happens to send —
		// this tells us whether cookie auth could be required on this
		// surface in the future.
		e := authlog.Start("dms", req.Request, nil)
		e.Note("exempt from auth")
		markPresented(e, req.Request)
		e.Done()
		chain.ProcessFilter(req, resp)
		return
	}

	if !common.IsUIAuthEnabled() || common.EnvConfig.NoAPIAuth {
		e := authlog.Start("api", req.Request, nil)
		e.Note("auth disabled")
		markPresented(e, req.Request)
		e.Done()
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
	e := authlog.Start("api", req.Request, rawBody)
	defer e.Done()
	markPresented(e, req.Request)

	// DeoVR deeplinks: DeoVR fetches /api/deovr/<scene-id>.json via a plain
	// GET and can present neither Basic credentials nor (reliably) the
	// player-session cookie, so the deep link carries a per-scene ?token=
	// derived from the player credentials and the scene ID. A match
	// authenticates exactly that one scene JSON; anything else falls through
	// to the normal Basic/cookie chain (which keeps the link working for the
	// owner in an authenticated browser session).
	if sceneID, ok := deoVRDeeplinkSceneID(req.Request.URL.Path); ok {
		if want := config.DeoVRDeeplinkToken(sceneID); want != "" {
			if got := req.Request.URL.Query().Get("token"); got != "" {
				if subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1 {
					e.AuthMethod = "deeplink-token"
					e.AuthResult = "accepted"
					chain.ProcessFilter(req, resp)
					return
				}
				e.Note("invalid deeplink token presented")
			}
		}
	}

	// CSRF note for the cookie path: unlike Basic Auth credentials, cookies
	// are attached by the browser to cross-site requests. Players are
	// non-standard browsers whose Origin headers cannot be relied upon, so
	// a mismatch is logged but the request is allowed. SameSite=Lax on the
	// cookie remains the effective layer in standard browsers.
	if isMutatingMethod(req.Request.Method) && !browserOriginMatchesHost(req.Request) {
		log.Warnf("mutating API request %s %s with non-matching Origin %q (Host %q) — allowing (non-standard browser)",
			req.Request.Method, req.Request.URL.Path, req.Request.Header.Get("Origin"), req.Request.Host)
		e.Note("origin mismatch on mutating request (allowed): Origin %q vs Host %q",
			req.Request.Header.Get("Origin"), req.Request.Host)
	}

	if hasValidPlayerSession(req.Request) {
		e.AuthMethod = "cookie"
		e.AuthResult = "accepted"
		chain.ProcessFilter(req, resp)
		return
	}

	user, password, ok := req.Request.BasicAuth()
	basicUI := ok && user == common.EnvConfig.UIUsername && checkUIPassword(password)
	basicPlayer := !basicUI && checkPlayerBasicAuth(user, password)
	switch {
	case !ok:
		e.AuthMethod = "none"
		e.AuthResult = "denied"
	case basicUI:
		e.AuthMethod = "basic-ui"
		e.AuthUser = user
		e.AuthResult = "success"
	case basicPlayer:
		e.AuthMethod = "basic-player"
		e.AuthUser = user
		e.AuthResult = "success"
	default:
		e.AuthMethod = "basic"
		e.AuthUser = user
		e.AuthResult = "failed"
	}
	if !basicUI && !basicPlayer {
		resp.AddHeader("WWW-Authenticate", `Basic realm="default"`)
		resp.WriteErrorString(http.StatusUnauthorized, "401: Unauthorized")
		return
	}
	if basicPlayer {
		// Mint the player-session cookie so subsequent requests from this
		// browser authenticate without the Basic prompt.
		setPlayerSessionCookie(e, resp.ResponseWriter)
	}

	chain.ProcessFilter(req, resp)
}

// deoVRDeeplinkSceneID extracts the numeric scene ID from a DeoVR deeplink
// JSON path (/api/deovr/<scene-id>.json). Returns ok=false for any other path
// or a malformed segment — those fall through to the normal auth chain.
func deoVRDeeplinkSceneID(path string) (uint, bool) {
	if !strings.HasPrefix(path, "/api/deovr/") {
		return 0, false
	}
	segment := strings.TrimPrefix(path, "/api/deovr/")
	if !strings.HasSuffix(segment, ".json") {
		return 0, false
	}
	id, err := strconv.ParseUint(strings.TrimSuffix(segment, ".json"), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
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
// http.ResponseWriter (Web UI surfaces) and marks the auth log entry.
func setPlayerSessionCookie(e *authlog.Entry, w http.ResponseWriter) {
	cookie := config.PlayerSessionCookie()
	if cookie == nil {
		return
	}
	http.SetCookie(w, cookie)
	e.CookieMinted = true
}

// markPresented records any authentication material the request carries
// (regardless of whether it will be accepted) on the auth log entry — used
// to learn whether players could authenticate on surfaces where auth is not
// (currently) required.
func markPresented(e *authlog.Entry, r *http.Request) {
	if user, _, ok := r.BasicAuth(); ok {
		e.PresentedBasicUser = user
	}
	if _, err := r.Cookie(config.PlayerSessionCookieName); err == nil {
		if hasValidPlayerSession(r) {
			e.PresentedPlayerCookie = "valid"
		} else {
			e.PresentedPlayerCookie = "invalid"
		}
	}
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

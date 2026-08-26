package server

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/xbapps/xbvr/pkg/authlog"
)

// Player detection: a request is treated as coming from a known VR player
// when its User-Agent matches one of the key-part regexes below, or when
// X-Requested-With exactly matches a known player package name. Either
// signal alone makes the client eligible for the player login page.
var playerUAPatterns = []*regexp.Regexp{
	regexp.MustCompile(`com\.heresphere\.vrvideoplayer`), // HereSphere media stack (Cronet)
	regexp.MustCompile(`HereSphere/[0-9]`),               // HereSphere API stack
	regexp.MustCompile(`\[DEO[0-9]`),                     // DeoVR, e.g. [DEO15.8.3888]
	regexp.MustCompile(`ExoPlayerLib`),                   // DeoVR/DVRHyper playback stack (AVPro)
	regexp.MustCompile(`\[HYPER[0-9]`),                   // DVRHyper, e.g. [HYPER1.12.2092]
}

var playerRequestedWith = []string{
	"com.heresphere.vrvideoplayer",
	"com.DeoVR.Hyper",
}

// isPlayerClient reports whether the request looks like it comes from a
// known VR player, by User-Agent regex or X-Requested-With match.
func isPlayerClient(r *http.Request) bool {
	ua := r.UserAgent()
	for _, p := range playerUAPatterns {
		if p.MatchString(ua) {
			return true
		}
	}
	rw := r.Header.Get("X-Requested-With")
	for _, known := range playerRequestedWith {
		if strings.EqualFold(rw, known) {
			return true
		}
	}
	return false
}

// loginPage is deliberately non-descript: no branding, just the form.
const loginPage = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Login</title>
</head>
<body>
<form method="POST" action="/login">
<input type="text" name="username" autocomplete="username" placeholder="Username">
<input type="password" name="password" autocomplete="current-password" placeholder="Password">
<button type="submit">Login</button>
</form>
</body>
</html>
`

func serveLoginPage(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	io.WriteString(w, loginPage)
}

// loginHandler serves the player login page (GET) and validates player
// credentials (POST). Only requests from known VR players may use it; all
// other clients get 403. A successful login sets the xbvr_player_session
// cookie and redirects to /ui/.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if !isPlayerClient(r) {
		authlog.Event("login", "%s %s from %s rejected (403, not a known player)", r.Method, r.URL.Path, r.RemoteAddr)
		http.Error(w, "403: Forbidden", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		authlog.Request("login", r, nil)
		if hasValidPlayerSession(r) {
			authlog.Event("login", "valid session cookie present, redirecting to /ui/")
			http.Redirect(w, r, "/ui/", http.StatusSeeOther)
			return
		}
		serveLoginPage(w, http.StatusOK)

	case http.MethodPost:
		// Buffer the form body for the auth log, then restore for ParseForm.
		var rawBody []byte
		if r.Body != nil && r.Body != http.NoBody {
			rawBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(rawBody))
		}
		authlog.Request("login", r, rawBody)

		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		if checkPlayerBasicAuth(username, password) {
			authlog.Event("login", "auth user=%q result=success, redirecting to /ui/", username)
			setPlayerSessionCookie("login", w)
			http.Redirect(w, r, "/ui/", http.StatusSeeOther)
			return
		}
		authlog.Event("login", "auth user=%q result=failed", username)
		serveLoginPage(w, http.StatusUnauthorized)

	default:
		http.Error(w, "405: Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

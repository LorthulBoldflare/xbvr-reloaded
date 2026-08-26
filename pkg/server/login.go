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
// Styled to match the main UI's design tokens, with light/dark support.
const loginPage = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark">
<title>Login</title>
<style>
  :root {
    --bg: #f4f5f8; --surface: #ffffff; --border: #cdd2dc;
    --text: #1c2333; --muted: #64708a; --primary: #4f46e5; --primary-strong: #4338ca;
    color-scheme: light;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #0d1017; --surface: #171c28; --border: #3a4358;
      --text: #e6e9f2; --muted: #9aa5bd; --primary: #818cf8; --primary-strong: #a5b0fc;
      color-scheme: dark;
    }
  }
  * { box-sizing: border-box; }
  body {
    margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center;
    background: var(--bg); color: var(--text);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
  }
  form {
    display: flex; flex-direction: column; gap: 0.8rem;
    width: min(340px, calc(100vw - 2rem));
    background: var(--surface); border: 1px solid var(--border);
    border-radius: 16px; padding: 1.5rem;
    box-shadow: 0 4px 12px rgba(16, 24, 40, 0.10);
  }
  input {
    width: 100%; padding: 0.6rem 0.75rem; font: inherit;
    color: var(--text); background: var(--surface);
    border: 1px solid var(--border); border-radius: 8px;
  }
  input::placeholder { color: var(--muted); }
  input:focus-visible {
    outline: none; border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.28);
  }
  button {
    margin-top: 0.25rem; padding: 0.6rem 0.75rem; font: inherit; font-weight: 600;
    color: #fff; background: var(--primary); border: none; border-radius: 8px; cursor: pointer;
  }
  button:hover { background: var(--primary-strong); }
  button:focus-visible { outline: none; box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.28); }
  @media (prefers-reduced-motion: no-preference) {
    input, button { transition: border-color 140ms, box-shadow 140ms, background-color 140ms; }
  }
</style>
</head>
<body>
<form method="POST" action="/login">
<input type="text" name="username" autocomplete="username" placeholder="Username" required autofocus>
<input type="password" name="password" autocomplete="current-password" placeholder="Password" required>
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
		e := authlog.Start("login", r, nil)
		e.AuthResult = "denied"
		e.Note("not a known player (403)")
		e.Done()
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		e := authlog.Start("login", r, nil)
		defer e.Done()
		e.PlayerClient = true
		if hasValidPlayerSession(r) {
			e.AuthMethod = "cookie"
			e.AuthResult = "accepted"
			e.RedirectTo = "/ui/"
			http.Redirect(w, r, "/ui/", http.StatusSeeOther)
			return
		}
		e.Note("served login form")
		serveLoginPage(w, http.StatusOK)

	case http.MethodPost:
		// Buffer the form body for the auth log, then restore for ParseForm.
		var rawBody []byte
		if r.Body != nil && r.Body != http.NoBody {
			rawBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(rawBody))
		}
		e := authlog.Start("login", r, rawBody)
		defer e.Done()
		e.PlayerClient = true

		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		e.AuthMethod = "form"
		e.AuthUser = username
		if checkPlayerBasicAuth(username, password) {
			e.AuthResult = "success"
			e.RedirectTo = "/ui/"
			setPlayerSessionCookie(e, w)
			http.Redirect(w, r, "/ui/", http.StatusSeeOther)
			return
		}
		e.AuthResult = "failed"
		serveLoginPage(w, http.StatusUnauthorized)

	default:
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// PlayerSessionCookieName carries the stable credential-derived token
// (PlayerSessionToken) from a successful native player login to the Web UI
// surfaces (/ui/, /api/*, /ws/), where it is accepted as an alternative to
// HTTP Basic Auth.
const PlayerSessionCookieName = "xbvr_player_session"

// PlayerAuthEnabled reports whether the DeoVR/HereSphere player endpoints
// require authentication. Both players share the Interfaces.DeoVR credential
// pair; this mirrors the check the player API filters perform.
func PlayerAuthEnabled() bool {
	return Config.Interfaces.DeoVR.AuthEnabled &&
		Config.Interfaces.DeoVR.Username != "" &&
		Config.Interfaces.DeoVR.Password != ""
}

// PlayerSessionToken derives the stable session token that lets a successful
// native player login (DeoVR/HereSphere) also authenticate Web UI requests
// via cookie. Keyed by the stored bcrypt password hash, it is stable across
// restarts, rotates automatically when either credential changes, and reveals
// neither the password nor (beyond the config file itself) the hash. Returns
// "" when player auth is disabled. There is deliberately no expiry or
// revocation: rotating the player password is the revocation mechanism.
func PlayerSessionToken() string {
	if !PlayerAuthEnabled() {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(Config.Interfaces.DeoVR.Password))
	mac.Write([]byte("xbvr-player-ui:" + Config.Interfaces.DeoVR.Username))
	return hex.EncodeToString(mac.Sum(nil))
}

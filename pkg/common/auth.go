package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func IsUIAuthEnabled() bool {
	if EnvConfig.UIUsername != "" && EnvConfig.UIPassword != "" {
		return true
	} else {
		return false
	}
}

// MCPToken derives the bearer token accepted by the /mcp endpoint from the
// configured UI credentials. The token is a keyed hash — it is stable across
// restarts while the credentials stay the same, rotates automatically when
// either changes, and (unlike the raw password) is safe to store in MCP
// client configs: leaking it does not reveal the UI password. Returns ""
// when UI auth is disabled (in which case /mcp is not served at all).
func MCPToken() string {
	if !IsUIAuthEnabled() {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(EnvConfig.UIPassword))
	mac.Write([]byte("xbvr-mcp:" + EnvConfig.UIUsername))
	return hex.EncodeToString(mac.Sum(nil))
}

func GetUISecret(user string, realm string) string {
	if user == EnvConfig.UIUsername {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(EnvConfig.UIPassword), bcrypt.DefaultCost)
		if err == nil {
			return string(hashedPassword)
		}
	}
	return ""
}

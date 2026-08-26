package config

import (
	"encoding/hex"
	"testing"
)

func TestPlayerSessionToken(t *testing.T) {
	orig := Config.Interfaces.DeoVR
	defer func() { Config.Interfaces.DeoVR = orig }()

	setCreds := func(enabled bool, user, pass string) {
		Config.Interfaces.DeoVR.AuthEnabled = enabled
		Config.Interfaces.DeoVR.Username = user
		Config.Interfaces.DeoVR.Password = pass
	}

	t.Run("empty when player auth disabled", func(t *testing.T) {
		setCreds(false, "UserA", "$2a$10$somebcrypthashvalue")
		if got := PlayerSessionToken(); got != "" {
			t.Fatalf("PlayerSessionToken() = %q, want empty (auth disabled)", got)
		}
		setCreds(true, "", "$2a$10$somebcrypthashvalue")
		if got := PlayerSessionToken(); got != "" {
			t.Fatalf("PlayerSessionToken() = %q, want empty (no username)", got)
		}
		setCreds(true, "UserA", "")
		if got := PlayerSessionToken(); got != "" {
			t.Fatalf("PlayerSessionToken() = %q, want empty (no password)", got)
		}
	})

	setCreds(true, "UserA", "$2a$10$somebcrypthashvalue")

	t.Run("deterministic and well-formed", func(t *testing.T) {
		token := PlayerSessionToken()
		if token != PlayerSessionToken() {
			t.Fatal("PlayerSessionToken() not deterministic")
		}
		decoded, err := hex.DecodeString(token)
		if err != nil || len(decoded) != 32 {
			t.Fatalf("PlayerSessionToken() = %q, want 64-char hex (32 bytes)", token)
		}
	})

	t.Run("rotates with password", func(t *testing.T) {
		before := PlayerSessionToken()
		Config.Interfaces.DeoVR.Password = "$2a$10$anotherbcrypthashvalue"
		defer setCreds(true, "UserA", "$2a$10$somebcrypthashvalue")
		if PlayerSessionToken() == before {
			t.Fatal("PlayerSessionToken() did not change after password change")
		}
	})

	t.Run("rotates with username", func(t *testing.T) {
		before := PlayerSessionToken()
		Config.Interfaces.DeoVR.Username = "UserB"
		defer setCreds(true, "UserA", "$2a$10$somebcrypthashvalue")
		if PlayerSessionToken() == before {
			t.Fatal("PlayerSessionToken() did not change after username change")
		}
	})
}

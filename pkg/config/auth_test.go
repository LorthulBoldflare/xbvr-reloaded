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

func TestDeoVRDeeplinkToken(t *testing.T) {
	orig := Config.Interfaces.DeoVR
	defer func() { Config.Interfaces.DeoVR = orig }()

	setCreds := func(enabled bool, user, pass string) {
		Config.Interfaces.DeoVR.AuthEnabled = enabled
		Config.Interfaces.DeoVR.Username = user
		Config.Interfaces.DeoVR.Password = pass
	}

	t.Run("empty when player auth disabled", func(t *testing.T) {
		setCreds(false, "UserA", "$2a$10$somebcrypthashvalue")
		if got := DeoVRDeeplinkToken(42); got != "" {
			t.Fatalf("DeoVRDeeplinkToken(42) = %q, want empty (auth disabled)", got)
		}
		setCreds(true, "", "$2a$10$somebcrypthashvalue")
		if got := DeoVRDeeplinkToken(42); got != "" {
			t.Fatalf("DeoVRDeeplinkToken(42) = %q, want empty (no username)", got)
		}
		setCreds(true, "UserA", "")
		if got := DeoVRDeeplinkToken(42); got != "" {
			t.Fatalf("DeoVRDeeplinkToken(42) = %q, want empty (no password)", got)
		}
	})

	setCreds(true, "UserA", "$2a$10$somebcrypthashvalue")

	t.Run("deterministic and well-formed", func(t *testing.T) {
		token := DeoVRDeeplinkToken(42)
		if token != DeoVRDeeplinkToken(42) {
			t.Fatal("DeoVRDeeplinkToken() not deterministic")
		}
		decoded, err := hex.DecodeString(token)
		if err != nil || len(decoded) != 32 {
			t.Fatalf("DeoVRDeeplinkToken() = %q, want 64-char hex (32 bytes)", token)
		}
	})

	t.Run("scoped per scene", func(t *testing.T) {
		if DeoVRDeeplinkToken(42) == DeoVRDeeplinkToken(43) {
			t.Fatal("DeoVRDeeplinkToken() must differ between scene IDs")
		}
	})

	t.Run("domain-separated from session token", func(t *testing.T) {
		if DeoVRDeeplinkToken(42) == PlayerSessionToken() {
			t.Fatal("DeoVRDeeplinkToken() must differ from PlayerSessionToken()")
		}
	})

	t.Run("rotates with credentials", func(t *testing.T) {
		before := DeoVRDeeplinkToken(42)
		Config.Interfaces.DeoVR.Password = "$2a$10$anotherbcrypthashvalue"
		if DeoVRDeeplinkToken(42) == before {
			t.Fatal("DeoVRDeeplinkToken() did not change after password change")
		}
		Config.Interfaces.DeoVR.Password = "$2a$10$somebcrypthashvalue"
		Config.Interfaces.DeoVR.Username = "UserB"
		if DeoVRDeeplinkToken(42) == before {
			t.Fatal("DeoVRDeeplinkToken() did not change after username change")
		}
	})
}

func TestNormalizePublicURL(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"  ", ""},
		{"https://my.xbvr.reloaded", "https://my.xbvr.reloaded"},
		{"https://my.xbvr.reloaded/", "https://my.xbvr.reloaded"},
		{" https://my.xbvr.reloaded/// ", "https://my.xbvr.reloaded"},
		{"http://192.168.1.10:9999/", "http://192.168.1.10:9999"},
		// invalid values disable the feature instead of half-enabling it
		{"https://", ""},
		{"http://", ""},
		{"my.xbvr.reloaded", ""},   // scheme-less
		{"localhost:9999", ""},     // parses as scheme "localhost"
		{"ftp://example.com", ""},  // non-http(s) scheme
	}
	for _, tt := range tests {
		if got := NormalizePublicURL(tt.in); got != tt.want {
			t.Errorf("NormalizePublicURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

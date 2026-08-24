package common

import (
	"encoding/hex"
	"testing"
)

func TestMCPToken(t *testing.T) {
	origUser, origPass := EnvConfig.UIUsername, EnvConfig.UIPassword
	defer func() { EnvConfig.UIUsername, EnvConfig.UIPassword = origUser, origPass }()

	t.Run("empty when UI auth disabled", func(t *testing.T) {
		EnvConfig.UIUsername, EnvConfig.UIPassword = "", ""
		if got := MCPToken(); got != "" {
			t.Fatalf("MCPToken() = %q, want empty", got)
		}
		EnvConfig.UIUsername, EnvConfig.UIPassword = "UserA", ""
		if got := MCPToken(); got != "" {
			t.Fatalf("MCPToken() = %q, want empty", got)
		}
	})

	EnvConfig.UIUsername, EnvConfig.UIPassword = "UserA", "Password123"

	t.Run("deterministic and well-formed", func(t *testing.T) {
		token := MCPToken()
		if token != MCPToken() {
			t.Fatal("MCPToken() not deterministic")
		}
		decoded, err := hex.DecodeString(token)
		if err != nil || len(decoded) != 32 {
			t.Fatalf("MCPToken() = %q, want 64-char hex (32 bytes)", token)
		}
	})

	t.Run("is not the raw credentials", func(t *testing.T) {
		if MCPToken() == EnvConfig.UIUsername+EnvConfig.UIPassword {
			t.Fatal("MCPToken() must not equal the concatenated credentials")
		}
	})

	t.Run("rotates with password", func(t *testing.T) {
		before := MCPToken()
		EnvConfig.UIPassword = "OtherPassword"
		defer func() { EnvConfig.UIPassword = "Password123" }()
		if MCPToken() == before {
			t.Fatal("MCPToken() did not change after password change")
		}
	})

	t.Run("rotates with username", func(t *testing.T) {
		before := MCPToken()
		EnvConfig.UIUsername = "UserB"
		defer func() { EnvConfig.UIUsername = "UserA" }()
		if MCPToken() == before {
			t.Fatal("MCPToken() did not change after username change")
		}
	})
}

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/xbapps/xbvr/pkg/common"
)

func runMCPAuth(t *testing.T, authHeader string) *httptest.ResponseRecorder {
	t.Helper()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	mcpAuthMiddleware(next).ServeHTTP(rec, req)
	return rec
}

func TestMCPAuthMiddleware(t *testing.T) {
	origUser, origPass := common.EnvConfig.UIUsername, common.EnvConfig.UIPassword
	defer func() {
		common.EnvConfig.UIUsername, common.EnvConfig.UIPassword = origUser, origPass
	}()

	common.EnvConfig.UIUsername = "UserA"
	common.EnvConfig.UIPassword = "Password123"

	t.Run("rejects missing token", func(t *testing.T) {
		rec := runMCPAuth(t, "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("rejects wrong token", func(t *testing.T) {
		rec := runMCPAuth(t, "Bearer UserAWrongPassword")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("rejects basic auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.SetBasicAuth("UserA", "Password123")
		rec := httptest.NewRecorder()
		mcpAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("accepts concatenated username+password token", func(t *testing.T) {
		rec := runMCPAuth(t, "Bearer UserAPassword123")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("open when UI auth disabled", func(t *testing.T) {
		common.EnvConfig.UIUsername = ""
		common.EnvConfig.UIPassword = ""
		defer func() { common.EnvConfig.UIUsername, common.EnvConfig.UIPassword = "UserA", "Password123" }()
		rec := runMCPAuth(t, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}

// TestMCPServerHandshake connects an MCP client to the streamable HTTP
// handler and verifies the three expected tools are registered.
func TestMCPServerHandshake(t *testing.T) {
	mcpServer := newMCPServer("test")
	ts := httptest.NewServer(mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return mcpServer }, nil))
	defer ts.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             ts.URL,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer session.Close()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools failed: %v", err)
	}

	got := map[string]bool{}
	for _, tool := range result.Tools {
		got[tool.Name] = true
	}
	for _, want := range []string{"rescan_storage", "scrape_scene", "generate_previews"} {
		if !got[want] {
			t.Fatalf("expected tool %q to be registered, got %v", want, got)
		}
	}
}

// TestMCPEndpointAuth verifies the bearer-token requirement end to end: an
// MCP client without a token is rejected, one with the concatenated
// username+password token completes the handshake.
func TestMCPEndpointAuth(t *testing.T) {
	origUser, origPass := common.EnvConfig.UIUsername, common.EnvConfig.UIPassword
	defer func() {
		common.EnvConfig.UIUsername, common.EnvConfig.UIPassword = origUser, origPass
	}()
	common.EnvConfig.UIUsername = "UserA"
	common.EnvConfig.UIPassword = "Password123"

	mcpServer := newMCPServer("test")
	ts := httptest.NewServer(mcpAuthMiddleware(mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return mcpServer }, nil)))
	defer ts.Close()

	ctx := context.Background()

	noAuthClient := mcp.NewClient(&mcp.Implementation{Name: "no-auth", Version: "0"}, nil)
	noAuthSession, err := noAuthClient.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             ts.URL,
		DisableStandaloneSSE: true,
	}, nil)
	if err == nil {
		noAuthSession.Close()
		t.Fatal("expected connect without token to fail")
	}

	authClient := mcp.NewClient(&mcp.Implementation{Name: "auth", Version: "0"}, nil)
	authSession, err := authClient.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint: ts.URL,
		HTTPClient: &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				r.Header.Set("Authorization", "Bearer UserAPassword123")
				return http.DefaultTransport.RoundTrip(r)
			}),
		},
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("expected connect with token to succeed: %v", err)
	}
	authSession.Close()
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

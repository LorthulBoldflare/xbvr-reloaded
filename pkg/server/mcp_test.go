package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
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
	for _, want := range []string{"rescan_storage", "scrape_scene", "generate_previews", "match_file"} {
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

// TestMCPMatchFile exercises the match_file tool handler against the
// throwaway sqlite DB that the models package init points at a temporary
// app dir when running under `go test`.
func TestMCPMatchFile(t *testing.T) {
	db, err := models.GetDB()
	if err != nil {
		t.Fatalf("get db: %v", err)
	}
	if err := db.AutoMigrate(&models.Scene{}, &models.Tag{}, &models.Actor{}, &models.File{},
		&models.Volume{}, &models.Action{}, &models.History{}, &models.SceneCuepoint{}).Error; err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	dir := t.TempDir()

	vol := models.Volume{Type: "local", Path: dir, IsEnabled: true}
	if err := db.Create(&vol).Error; err != nil {
		t.Fatal(err)
	}

	scene := models.Scene{SceneID: "mcptest-scene-1", Title: "MCP Test Scene", FilenamesArr: "[]"}
	if err := db.Create(&scene).Error; err != nil {
		t.Fatal(err)
	}

	otherScene := models.Scene{SceneID: "mcptest-scene-2", Title: "Other Scene", FilenamesArr: "[]"}
	if err := db.Create(&otherScene).Error; err != nil {
		t.Fatal(err)
	}

	newFile := func(name string) models.File {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		f := models.File{Filename: name, Path: dir, VolumeID: vol.ID, Type: "video"}
		if err := db.Create(&f).Error; err != nil {
			t.Fatal(err)
		}
		return f
	}

	call := func(filename, sceneID string) (*mcp.CallToolResult, error) {
		t.Helper()
		res, _, err := mcpMatchFile(context.Background(), nil, mcpMatchFileArgs{Filename: filename, SceneID: sceneID})
		return res, err
	}

	t.Run("matches file to scene", func(t *testing.T) {
		f := newFile("mcptest-match.mp4")
		res, err := call(f.Filename, scene.SceneID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := res.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(text, "Matched") || !strings.Contains(text, scene.SceneID) {
			t.Fatalf("unexpected result text: %q", text)
		}

		var updated models.File
		db.First(&updated, f.ID)
		if updated.SceneID != scene.ID {
			t.Fatalf("file not assigned to scene: scene_id = %d, want %d", updated.SceneID, scene.ID)
		}

		var updatedScene models.Scene
		db.First(&updatedScene, scene.ID)
		if !strings.Contains(updatedScene.FilenamesArr, f.Filename) {
			t.Fatalf("filename not added to scene filenames_arr: %q", updatedScene.FilenamesArr)
		}

		var actions []models.Action
		db.Where("scene_id = ? AND action_type = ?", scene.SceneID, "match").Find(&actions)
		if len(actions) == 0 {
			t.Fatal("expected a match action to be recorded")
		}

		t.Run("repeat call is idempotent", func(t *testing.T) {
			res, err := call(f.Filename, scene.SceneID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "already matched") {
				t.Fatalf("unexpected result text: %q", res.Content[0].(*mcp.TextContent).Text)
			}
		})
	})

	t.Run("unknown scene id errors", func(t *testing.T) {
		_, err := call("mcptest-match.mp4", "mcptest-no-such-scene")
		if err == nil {
			t.Fatal("expected error for unknown scene_id")
		}
	})

	t.Run("unknown filename errors", func(t *testing.T) {
		_, err := call("mcptest-not-on-disk.mp4", scene.SceneID)
		if err == nil {
			t.Fatal("expected error for unknown filename")
		}
	})

	t.Run("conflicting match errors", func(t *testing.T) {
		f := newFile("mcptest-taken.mp4")
		f.SceneID = otherScene.ID
		if err := db.Save(&f).Error; err != nil {
			t.Fatal(err)
		}
		_, err := call(f.Filename, scene.SceneID)
		if err == nil {
			t.Fatal("expected error for file matched to a different scene")
		}
	})

	t.Run("duplicate scene id errors", func(t *testing.T) {
		dup := models.Scene{SceneID: "mcptest-dup", Title: "Dup A", FilenamesArr: "[]"}
		if err := db.Create(&dup).Error; err != nil {
			t.Fatal(err)
		}
		dup2 := models.Scene{SceneID: "mcptest-dup", Title: "Dup B", FilenamesArr: "[]"}
		if err := db.Create(&dup2).Error; err != nil {
			t.Fatal(err)
		}
		f := newFile("mcptest-dupscene.mp4")
		_, err := call(f.Filename, "mcptest-dup")
		if err == nil {
			t.Fatal("expected error for ambiguous scene_id")
		}
	})
}

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/jinzhu/gorm"

	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/models"
)

// setupAdvancedSaveDB swaps in an in-memory DB with the kvs table so
// config.SaveConfig can persist.
func setupAdvancedSaveDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open("sqlite3", "file:"+strings.ReplaceAll(t.Name(), "/", "-")+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	db.DB().SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.KV{}).Error; err != nil {
		t.Fatal(err)
	}

	restore := models.SetCommonDBForTests(db)
	t.Cleanup(func() {
		restore()
		db.Close()
	})
}

func saveAdvanced(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()

	container := restful.NewContainer()
	container.Add(ConfigResource{}.WebService())

	req := httptest.NewRequest(http.MethodPut, "http://xbvr.local/api/options/interface/advanced", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	container.Dispatch(rec, req)
	return rec
}

func TestSaveOptionsAdvancedDateFormats(t *testing.T) {
	setupAdvancedSaveDB(t)

	origAdvanced := config.Config.Advanced
	origPublic := config.Config.Server.PublicURL
	t.Cleanup(func() {
		config.Config.Advanced = origAdvanced
		config.Config.Server.PublicURL = origPublic
	})

	cases := []struct {
		name     string
		body     string
		wantDate time.Time
	}{
		{
			name:     "web UI date-only payload",
			body:     `{"stashApiKey":"","ignoreReleasedBefore":"2026-08-28","publicUrl":"https://example.com/"}`,
			wantDate: time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "old UI RFC3339 payload",
			body:     `{"stashApiKey":"","ignoreReleasedBefore":"2026-08-28T10:00:00Z","publicUrl":"https://example.com"}`,
			wantDate: time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "old UI null date clears",
			body:     `{"stashApiKey":"","ignoreReleasedBefore":null,"publicUrl":"https://example.com"}`,
			wantDate: time.Time{},
		},
		{
			name:     "web UI empty date clears",
			body:     `{"stashApiKey":"","ignoreReleasedBefore":"","publicUrl":"https://example.com"}`,
			wantDate: time.Time{},
		},
		{
			name:     "absent publicUrl leaves stored value unchanged",
			body:     `{"stashApiKey":"","ignoreReleasedBefore":"2026-08-28"}`,
			wantDate: time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config.Config.Server.PublicURL = "https://previously.stored/"
			config.Config.Advanced.IgnoreReleasedBefore = time.Time{}

			rec := saveAdvanced(t, tc.body)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}

			// The whole save must go through: a decode failure on
			// ignoreReleasedBefore used to abort the handler before
			// publicUrl was applied.
			if !config.Config.Advanced.IgnoreReleasedBefore.Equal(tc.wantDate) {
				t.Errorf("IgnoreReleasedBefore = %v, want %v", config.Config.Advanced.IgnoreReleasedBefore, tc.wantDate)
			}

			wantPublic := "https://previously.stored/"
			if strings.Contains(tc.body, "publicUrl") {
				wantPublic = "https://example.com"
			}
			if config.Config.Server.PublicURL != wantPublic {
				t.Errorf("PublicURL = %q, want %q", config.Config.Server.PublicURL, wantPublic)
			}
		})
	}
}

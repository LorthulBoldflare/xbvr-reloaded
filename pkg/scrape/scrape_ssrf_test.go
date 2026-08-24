package scrape

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCreateCollectorBlocksLoopback proves the collectors built by
// createCollector route every request through common.SSRFSafeTransport: a
// page fetch to a loopback address (as an httptest server always is) must be
// rejected before any network traffic happens. This closes the
// validate-then-fetch gap where a URL validated once as public could
// re-resolve to, or redirect to, an internal address at fetch time.
func TestCreateCollectorBlocksLoopback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := createCollector("127.0.0.1", "localhost", "::1")
	err := c.Visit(ts.URL)
	if err == nil {
		t.Fatal("expected collector visit to loopback to be blocked")
	}
	if !strings.Contains(err.Error(), "outbound request blocked") {
		t.Fatalf("expected SSRF block error, got: %v", err)
	}
}

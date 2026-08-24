package common

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateOutboundURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		// public targets
		{"https://stashdb.org/graphql", false},
		{"http://example.com/page", false},
		// internal targets — SSRF
		{"http://127.0.0.1:9999/api/options/state", true},
		{"http://localhost/admin", true},
		{"http://169.254.169.254/latest/meta-data", true},
		{"http://10.0.0.1/", true},
		{"http://172.16.0.5/", true},
		{"http://192.168.1.1/", true},
		{"http://[::1]/", true},
		// malformed / wrong scheme
		{"ftp://example.com/file", true},
		{"file:///etc/passwd", true},
		{"://no-scheme", true},
		{"http://", true},
		{"", true},
	}
	for _, tt := range tests {
		err := ValidateOutboundURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateOutboundURL(%q) err = %v, wantErr %v", tt.url, err, tt.wantErr)
		}
	}
}

// TestSSRFSafeTransportBlocksPerHop verifies the transport re-validates every
// request it carries — http.Client calls RoundTrip once per redirect hop, so
// this is what stops a validated public URL from 302ing to an internal one.
func TestSSRFSafeTransportBlocksPerHop(t *testing.T) {
	loopback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer loopback.Close()

	req, err := http.NewRequest(http.MethodGet, loopback.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (SSRFSafeTransport{}).RoundTrip(req); err == nil {
		t.Fatal("expected RoundTrip to block loopback target")
	} else if !strings.Contains(err.Error(), "outbound request blocked") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Non-http(s) schemes never reach the network either.
	req, err = http.NewRequest(http.MethodGet, "file:///etc/passwd", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (SSRFSafeTransport{}).RoundTrip(req); err == nil {
		t.Fatal("expected RoundTrip to block file:// scheme")
	}
}

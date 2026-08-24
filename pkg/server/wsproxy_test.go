package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWsOriginAllowed(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		{"no origin header (non-browser client)", "192.168.1.10:9999", "", true},
		{"same origin", "192.168.1.10:9999", "http://192.168.1.10:9999", true},
		{"same origin https", "example.local:9999", "https://example.local:9999", true},
		{"cross origin (CSWSH attempt)", "192.168.1.10:9999", "http://evil.example.com", false},
		{"cross origin same host different port", "192.168.1.10:9999", "http://192.168.1.10:9998", false},
		{"invalid origin", "192.168.1.10:9999", "://not a url", false},
		{"subdomain is still cross-origin", "example.com:9999", "http://evil.example.com:9999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://"+tt.host+"/ws/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if got := wsOriginAllowed(req); got != tt.want {
				t.Fatalf("wsOriginAllowed(host=%q origin=%q) = %v, want %v", tt.host, tt.origin, got, tt.want)
			}
		})
	}
}

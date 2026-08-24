package common

import "testing"

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

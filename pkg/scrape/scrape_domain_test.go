package scrape

import "testing"

func TestGetCoreDomain(t *testing.T) {
	tests := []struct{ in, want string }{
		{"www.example.com", "example"},
		{"example.com", "example"},
		{"https://www.theporndb.net/some/page", "theporndb"},
		// multi-part TLD: previously returned "example.co"
		{"www.example.co.uk", "example"},
		{"example.co.uk", "example"},
		{"https://site.com.au/path", "site"},
		// unlisted/invalid domains fall back to the old heuristic
		{"scraper.local", "scraper"},
		{"localhost", "localhost"},
	}
	for _, tt := range tests {
		if got := GetCoreDomain(tt.in); got != tt.want {
			t.Errorf("GetCoreDomain(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

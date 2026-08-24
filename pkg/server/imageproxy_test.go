package server

import (
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"willnorris.com/go/imageproxy"
)

func newTestProxy() *imageproxy.Proxy {
	p := imageproxy.NewProxy(nil, nil)
	p.DenyHosts = deniedProxyHosts
	return p
}

func proxyRequest(p *imageproxy.Proxy, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/0x/"+target, nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	return rec
}

func TestImageProxySSRFProtection(t *testing.T) {
	p := newTestProxy()

	t.Run("rejects internal targets", func(t *testing.T) {
		targets := []string{
			"http:/169.254.169.254/latest/meta-data", // cloud metadata
			"http:/127.0.0.1:9999/api/options/state",
			"http:/localhost/admin",
			"http:/10.0.0.1/internal",
			"http:/172.16.0.1/internal",
			"http:/192.168.1.1/internal",
			"http:/100.64.0.1/internal",
		}
		for _, target := range targets {
			rec := proxyRequest(p, target)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("target %q: expected 403, got %d", target, rec.Code)
			}
		}
	})

	t.Run("rejects relative URLs without DefaultBaseURL", func(t *testing.T) {
		rec := proxyRequest(p, "/api/dms/file/1")
		if rec.Code == http.StatusOK {
			t.Fatalf("expected relative URL to be rejected, got 200")
		}
	})
}

func TestImageProxyLegitRemoteStillWorks(t *testing.T) {
	// Spin up an "upstream" image server. It binds 127.0.0.1, which the proxy
	// denies, so this test uses a proxy without DenyHosts to prove the proxy
	// + cache mechanics themselves are intact.
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, img, nil)
	}))
	defer upstream.Close()

	p := imageproxy.NewProxy(nil, nil)
	target := strings.Replace(upstream.URL, "://", ":/", 1)
	rec := proxyRequest(p, target)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/jpeg") {
		t.Fatalf("expected image/jpeg, got %q", ct)
	}
}

func TestDeniedProxyHostsCoversDocumentedRanges(t *testing.T) {
	required := []string{
		"127.0.0.0/8", "169.254.0.0/16", "10.0.0.0/8",
		"172.16.0.0/12", "192.168.0.0/16", "localhost",
	}
	for _, want := range required {
		found := false
		for _, h := range deniedProxyHosts {
			if h == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("deniedProxyHosts missing %q (entries: %s)", want, fmt.Sprint(deniedProxyHosts))
		}
	}
}

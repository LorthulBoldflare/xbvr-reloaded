package common

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// SSRFSafeTransport wraps an http.RoundTripper and re-validates the target of
// every request — including redirect hops, since http.Client calls RoundTrip
// per hop — so a validated URL cannot 302 to (or rebind to) an internal
// address. Use it on any client that fetches user-influenced URLs.
type SSRFSafeTransport struct {
	Base http.RoundTripper
}

func (t SSRFSafeTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := ValidateOutboundURL(r.URL.String()); err != nil {
		return nil, fmt.Errorf("outbound request blocked: %v", err)
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(r)
}

// ValidateOutboundURL checks a user-supplied URL that the server is about to
// fetch (scraping, trailer lookup, bundle download). Only public http(s)
// targets are allowed: the host must resolve, and none of its addresses may
// be loopback, link-local, private, unspecified or multicast. This blocks
// SSRF against internal services (cloud metadata, localhost admin panels,
// LAN devices).
func ValidateOutboundURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL scheme %q is not allowed, only http/https", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL has no host")
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("cannot resolve host %q: %v", host, err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
			ip.IsUnspecified() || ip.IsPrivate() || ip.IsMulticast() {
			return fmt.Errorf("host %q resolves to a non-public address (%s)", host, ip)
		}
	}
	return nil
}

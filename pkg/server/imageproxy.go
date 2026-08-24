package server

// deniedProxyHosts lists hosts and CIDR ranges the image proxy must never
// connect to: loopback, link-local (cloud metadata endpoints), unspecified,
// RFC1918/private and carrier-grade NAT ranges. Matching is performed by
// willnorris.com/go/imageproxy's hostMatches (exact host, *.suffix, or CIDR).
var deniedProxyHosts = []string{
	"localhost",
	"127.0.0.0/8",    // IPv4 loopback
	"::1/128",        // IPv6 loopback
	"0.0.0.0/8",      // IPv4 unspecified
	"169.254.0.0/16", // IPv4 link-local (incl. 169.254.169.254 cloud metadata)
	"fe80::/10",      // IPv6 link-local
	"10.0.0.0/8",     // RFC1918
	"172.16.0.0/12",  // RFC1918
	"192.168.0.0/16", // RFC1918
	"fd00::/8",       // IPv6 unique local
	"100.64.0.0/10",  // carrier-grade NAT
}

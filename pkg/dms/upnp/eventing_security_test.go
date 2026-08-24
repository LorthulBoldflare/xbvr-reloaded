package upnp

import (
	"net"
	"testing"
	"time"
)

func TestValidCallbackURLs(t *testing.T) {
	subscriber := net.ParseIP("192.168.1.50")

	urls := ParseCallbackURLs(
		"<http://192.168.1.50:8080/events>" + // the subscriber itself — OK
			"<http://192.168.1.99:9090/hook>" + // other LAN host — rejected
			"<http://127.0.0.1:9999/api/task/scrape>" + // loopback SSRF — rejected
			"<http://169.254.169.254/latest/meta-data>" + // metadata SSRF — rejected
			"<file:///etc/passwd>" + // non-http scheme — rejected
			"<https://192.168.1.50:8443/events>", // subscriber over https — OK
	)

	valid := ValidCallbackURLs(urls, subscriber)
	if len(valid) != 2 {
		t.Fatalf("expected 2 valid callback URLs, got %d: %v", len(valid), valid)
	}
	for _, u := range valid {
		if u.Hostname() != "192.168.1.50" {
			t.Fatalf("unexpected valid callback URL: %s", u)
		}
	}

	if got := ValidCallbackURLs(urls, nil); len(got) != 0 {
		t.Fatalf("expected all callbacks rejected for unparseable subscriber, got %v", got)
	}
}

func TestSubscribePrunesExpiredAndUnsubscribeWorks(t *testing.T) {
	var ev Eventing
	callback := ParseCallbackURLs("<http://192.168.1.50:8080/events>")

	// short-lived subscription that expires immediately
	sid1, _, err := ev.Subscribe(callback, 0)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	// subscribing again must prune the expired entry (map must not grow)
	sid2, timeout, err := ev.Subscribe(callback, 300)
	if err != nil {
		t.Fatal(err)
	}
	if timeout < 299 || timeout > 300 {
		t.Fatalf("expected timeout ~300, got %d", timeout)
	}
	ev.mu.Lock()
	remaining := len(ev.subscribers)
	ev.mu.Unlock()
	if remaining != 1 {
		t.Fatalf("expected 1 subscriber after pruning, got %d", remaining)
	}

	if err := ev.Unsubscribe(sid2); err != nil {
		t.Fatalf("Unsubscribe(%s): %v", sid2, err)
	}
	if err := ev.Unsubscribe(sid1); err == nil {
		t.Fatal("expected error unsubscribing unknown/expired sid")
	}
}

package ssdp

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestParseMX(t *testing.T) {
	tests := []struct {
		in      string
		want    uint
		wantErr bool
	}{
		{"1", 1, false},
		{"3", 3, false},
		{"5", 5, false},
		{"0", 1, false},           // clamped up
		{"1000000", maxMX, false}, // clamped down
		{"999999999999", maxMX, false},
		{"18446744073709551616", 0, true}, // out of uint64 range
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		got, err := parseMX(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseMX(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseMX(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestMXDelayNeverPanics(t *testing.T) {
	// regression: int64(time.Second) * int64(mx) overflowed negative for
	// huge MX values, panicking rand.Int63n in a per-packet goroutine and
	// killing the whole process
	for _, mxHeader := range []string{"1", "5", "1000000", "999999999999"} {
		mx, err := parseMX(mxHeader)
		if err != nil {
			t.Fatalf("parseMX(%q): %v", mxHeader, err)
		}
		delay := time.Duration(rand.Int63n(int64(time.Second) * int64(mx)))
		if delay < 0 || delay > maxMX*time.Second {
			t.Fatalf("mx %q produced out-of-range delay %v", mxHeader, delay)
		}
	}
}

func loopbackInterface(t *testing.T) net.Interface {
	t.Helper()
	for _, name := range []string{"lo0", "lo"} {
		if ifi, err := net.InterfaceByName(name); err == nil {
			return *ifi
		}
	}
	t.Skip("no loopback interface available")
	return net.Interface{}
}

// TestHandleFuzzedMSearch feeds crafted M-SEARCH packets (including huge MX
// values that used to crash the process) through handle and asserts no panic.
func TestHandleFuzzedMSearch(t *testing.T) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Skipf("cannot open UDP socket: %v", err)
	}
	defer conn.Close()

	srv := &Server{
		conn:           conn,
		Interface:      loopbackInterface(t),
		Server:         "test/1.0",
		UUID:           "uuid:test",
		NotifyInterval: time.Minute,
		closed:         make(chan struct{}),
		Location:       func(ip net.IP) string { return "http://127.0.0.1/rootDesc.xml" },
	}
	defer close(srv.closed)

	sender := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1900}

	mxValues := []string{"1", "5", "0", "1000000", "999999999999", "18446744073709551616", "abc", "-1"}
	for _, mx := range mxValues {
		t.Run(fmt.Sprintf("mx=%s", mx), func(t *testing.T) {
			pkt := fmt.Sprintf("M-SEARCH * HTTP/1.1\r\nHost: %s\r\nMan: \"ssdp:discover\"\r\nMx: %s\r\nSt: ssdp:all\r\n\r\n", AddrString, mx)
			srv.handle([]byte(pkt), sender) // must not panic
		})
	}
}

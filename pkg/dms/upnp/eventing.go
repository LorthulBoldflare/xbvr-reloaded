package upnp

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"regexp"
	"sync"
	"time"
)

// TODO: Why use namespace prefixes in PropertySet et al? Because the spec
// uses them, and I believe the Golang standard library XML spec implementers
// incorrectly assume that you can get away with just xmlns="".

// propertyset is the root element sent in an event callback.
type PropertySet struct {
	XMLName    struct{} `xml:"e:propertyset"`
	Properties []Property
	// This should be set to `"urn:schemas-upnp-org:event-1-0"`.
	Space string `xml:"xmlns:e,attr"`
}

// propertys provide namespacing to the contained variables.
type Property struct {
	XMLName  struct{} `xml:"e:property"`
	Variable Variable
}

// Represents an evented state variable that has sendEvents="yes" in its
// service spec.
type Variable struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type subscriber struct {
	sid     string
	nextSeq uint32 // 0 for initial event, wraps from Uint32Max to 1.
	urls    []*url.URL
	expiry  time.Time
}

// Intended to eventually be an embeddable implementation for managing
// eventing for a service. Not complete.
type Eventing struct {
	mu          sync.Mutex
	subscribers map[string]*subscriber
}

func (me *Eventing) Subscribe(callback []*url.URL, timeoutSeconds int) (sid string, actualTimeout int, err error) {
	var uuid [16]byte
	io.ReadFull(rand.Reader, uuid[:])
	sid = FormatUUID(uuid[:])

	me.mu.Lock()
	defer me.mu.Unlock()
	me.pruneExpiredLocked()
	if _, ok := me.subscribers[sid]; ok {
		err = fmt.Errorf("already subscribed: %s", sid)
		return
	}
	if me.subscribers == nil {
		me.subscribers = make(map[string]*subscriber)
	}
	ssr := &subscriber{
		sid:    sid,
		urls:   callback,
		expiry: time.Now().Add(time.Duration(timeoutSeconds) * time.Second),
	}
	me.subscribers[sid] = ssr
	actualTimeout = int(ssr.expiry.Sub(time.Now()) / time.Second)
	return
}

// pruneExpiredLocked drops expired subscriptions so the subscriber map does
// not grow unboundedly from clients that never unsubscribe. Caller must hold
// me.mu.
func (me *Eventing) pruneExpiredLocked() {
	now := time.Now()
	for sid, sub := range me.subscribers {
		if now.After(sub.expiry) {
			delete(me.subscribers, sid)
		}
	}
}

func (me *Eventing) Unsubscribe(sid string) error {
	me.mu.Lock()
	defer me.mu.Unlock()
	if _, ok := me.subscribers[sid]; !ok {
		return fmt.Errorf("unknown subscription: %s", sid)
	}
	delete(me.subscribers, sid)
	return nil
}

var callbackURLRegexp = regexp.MustCompile("<(.*?)>")

// Parse the CALLBACK HTTP header in an event subscription request. See UPnP
// Device Architecture 4.1.2.
func ParseCallbackURLs(callback string) (ret []*url.URL) {
	for _, match := range callbackURLRegexp.FindAllStringSubmatch(callback, -1) {
		_url, err := url.Parse(match[1])
		if err != nil {
			log.Printf("bad callback url: %q", match[1])
			continue
		}
		ret = append(ret, _url)
	}
	return
}

// ValidCallbackURLs filters subscription callback URLs to those that are safe
// to NOTIFY: http(s) scheme, and targeting the subscribing control point
// itself (per UPnP DA the event callback is delivered back to the
// subscriber). This prevents SUBSCRIBE requests from being abused to make the
// server POST to arbitrary (e.g. loopback or link-local metadata) endpoints.
func ValidCallbackURLs(urls []*url.URL, subscriberIP net.IP) (ret []*url.URL) {
	if subscriberIP == nil {
		return nil
	}
	for _, u := range urls {
		if u.Scheme != "http" && u.Scheme != "https" {
			log.Printf("rejected event callback with scheme %q: %s", u.Scheme, u)
			continue
		}
		host := u.Hostname()
		if ip := net.ParseIP(host); ip != nil {
			if !ip.Equal(subscriberIP) {
				log.Printf("rejected event callback to %s: not the subscriber %s", ip, subscriberIP)
				continue
			}
			ret = append(ret, u)
			continue
		}
		// hostname callback: resolve and require the subscriber's address
		ips, err := net.LookupIP(host)
		if err != nil {
			log.Printf("rejected event callback with unresolvable host %q: %s", host, err)
			continue
		}
		ok := false
		for _, ip := range ips {
			if ip.Equal(subscriberIP) {
				ok = true
				break
			}
		}
		if !ok {
			log.Printf("rejected event callback to %q: does not resolve to subscriber %s", host, subscriberIP)
			continue
		}
		ret = append(ret, u)
	}
	return
}

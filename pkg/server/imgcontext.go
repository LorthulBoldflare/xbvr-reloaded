package server

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
)

// ImageProxyContextHandler wraps the image proxy with a leading "context"
// path segment: /img/<context>/<options>/<remote-url>.
//
//	context := "scene-"<scene_id>  → scene (scenes.scene_id; any non-empty
//	                                 slug, e.g. scene-slr-1234, scene-stash-<uuid>)
//	          | "act-"<id>         → actor (actors.id, numeric PK)
//	          | "icon-"<slug>      → icon ([a-z0-9]+, e.g. icon-baberotica)
//
// A literal 0 id ("scene-0" / "act-0" / "icon-0") means unattributed: the
// request is passed through to the proxy but never recorded. Requests whose
// context resolves to an existing scene/actor (icons always) are recorded in
// image_proxy_entries (strict policy: unknown ids proxy fine but write no
// row). Anything not matching the grammar above is rejected with 400.
//
// The context segment is plain unencoded ASCII in both Path and RawPath, so
// stripping it with http.StripPrefix leaves the remainder byte-identical to
// legacy /img/<options>/<url> requests — the outbound request imageproxy
// builds (and therefore the httpcache disk key) is unchanged and existing
// on-disk cache entries are reused. Do NOT re-encode the remainder.
//
// The options segment is passed through to imageproxy unmodified; UIs use the
// literal token "raw" for a full-size passthrough (imageproxy ignores unknown
// option tokens, so "raw" behaves like empty options and shares their
// "0x0" cache key).
//
// Two imageproxy backends are wired in: a caching one (disk cache,
// ForceCache, FollowRedirects) and a non-caching one (NopCache). Routing is
// purely syntactic: zero-id contexts ("scene-0"/"act-0"/"icon-0") are
// transient/unattributed and bypass the disk cache; every other valid
// context uses the caching backend — including unknown-but-valid ids, which
// preserves legacy cache behaviour. Invariant: recordable ⇔ cached (both
// gate on id != "0").
type ImageProxyContextHandler struct {
	caching http.Handler
	noCache http.Handler
	// recorded caches (context, url, options) triples already written this
	// process lifetime. Positive results only: a failed existence check must
	// never be cached, or a scene saved after its images were first requested
	// (normal scrape flow) would never be recorded until restart. Consistent
	// because the table is insert-only: rows are only ever added.
	recorded sync.Map
}

func NewImageProxyContextHandler(caching, noCache http.Handler) *ImageProxyContextHandler {
	return &ImageProxyContextHandler{caching: caching, noCache: noCache}
}

var (
	// imgPathRe splits /img/<context>/<options>/<url> in one pass. All three
	// segments must be non-empty — this also rejects "/img/<seg>" outright,
	// which would otherwise strip to an empty URL.Path and panic imageproxy's
	// NewRequest (EscapedPath()[1:]).
	imgPathRe = regexp.MustCompile(`^/img/([^/]+)/([^/]+)/(.+)$`)

	sceneContextRe = regexp.MustCompile(`^scene-([^/]+)$`)
	actorContextRe = regexp.MustCompile(`^act-([0-9]+)$`)
	iconContextRe  = regexp.MustCompile(`^icon-([a-z0-9]+)$`)

	// cleanedURLRe normalizes the remote URL exactly like imageproxy's own
	// reCleanedURL: both the ":/" scheme hack and the full "://" form resolve
	// to "://", never ":///".
	cleanedURLRe = regexp.MustCompile(`^(https?):/+([^/])`)
)

// parseImageContext validates and splits the context segment; ok=false means
// the segment does not match the established grammar.
func parseImageContext(seg string) (kind, id string, ok bool) {
	if m := sceneContextRe.FindStringSubmatch(seg); m != nil {
		return "scene", m[1], true
	}
	if m := actorContextRe.FindStringSubmatch(seg); m != nil {
		return "act", m[1], true
	}
	if m := iconContextRe.FindStringSubmatch(seg); m != nil {
		return "icon", m[1], true
	}
	return "", "", false
}

func (h *ImageProxyContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := imgPathRe.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.Error(w, "image proxy: malformed path, want /img/<context>/<options>/<url>", http.StatusBadRequest)
		return
	}
	seg, options, urlPart := m[1], m[2], m[3]

	kind, id, ok := parseImageContext(seg)
	if !ok {
		http.Error(w, "image proxy: invalid context, want scene-<id>|act-<id>|icon-<slug>", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.recordAssociation(seg, kind, id, options, urlPart, r.URL.RawQuery)
	}

	backend := h.caching
	if id == "0" {
		backend = h.noCache
	}
	http.StripPrefix("/img/"+seg, backend).ServeHTTP(w, r)
}

// recordAssociation best-effort records the image<->entity association for
// this request. It must never affect serving the image.
func (h *ImageProxyContextHandler) recordAssociation(seg, kind, id, options, urlPart, rawQuery string) {
	// literal 0 id = unattributed: pass-through, no recording
	if id == "0" {
		return
	}

	remoteURL := cleanedURLRe.ReplaceAllString(urlPart, "$1://$2")
	if !strings.HasPrefix(remoteURL, "http://") && !strings.HasPrefix(remoteURL, "https://") {
		return
	}
	// The remote url's query string, if any, arrived as this request's RawQuery.
	if rawQuery != "" {
		remoteURL += "?" + rawQuery
	}

	// Cheap gate before any DB work: already recorded this process lifetime.
	urlHash := models.ImageProxyURLHash(remoteURL, options)
	key := seg + "\x1f" + urlHash
	if _, ok := h.recorded.Load(key); ok {
		return
	}

	db, err := models.GetDB()
	if err != nil || db == nil {
		return
	}

	switch kind {
	case "scene":
		// soft-deleted scenes are excluded by gorm's default scope
		if !rowExists(db.Model(&models.Scene{}).Where("scene_id = ?", id)) {
			return
		}
	case "act":
		if !rowExists(db.Model(&models.Actor{}).Where("id = ?", id)) {
			return
		}
	case "icon":
		// icons have no backing entity; always record
	}

	models.RecordImageProxyEntry(db, seg, remoteURL, options)
	h.recorded.Store(key, struct{}{})
}

func rowExists(q *gorm.DB) bool {
	var count int
	if err := q.Count(&count).Error; err != nil {
		common.Log.Debugf("imageproxy: existence check failed: %v", err)
		return false
	}
	return count > 0
}

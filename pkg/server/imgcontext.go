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
//	context := "0"            → unattributed (never recorded)
//	          | "act-"<id>    → actor (actors.id; "act-0" never recorded)
//	          | "icon-"<slug> → icon (^[a-z0-9]+$, always recorded)
//	          | <scene_id>    → scenes.scene_id (any other non-empty segment)
//
// Requests whose context resolves to an existing scene/actor are recorded in
// image_proxy_entries (strict policy: unknown contexts proxy fine but write
// no row). UIs insert the context segment raw, matching how scene ids are
// already interpolated into /scenes/<scene_id> routes.
//
// The context segment is plain unencoded ASCII in both Path and RawPath, so
// stripping it with http.StripPrefix leaves the remainder byte-identical to
// legacy /img/<options>/<url> requests — the outbound request imageproxy
// builds (and therefore the httpcache disk key) is unchanged and existing
// on-disk cache entries are reused. Do NOT re-encode the remainder.
//
// Scene ids containing a literal "/" are unsupported (none exist today:
// scrapers build them from slugified site names and id components).
type ImageProxyContextHandler struct {
	proxy http.Handler
	// recorded caches (context, url, options) triples already written this
	// process lifetime. Positive results only: a failed existence check must
	// never be cached, or a scene saved after its images were first requested
	// (normal scrape flow) would never be recorded until restart. Consistent
	// because the table is insert-only: rows are only ever added.
	recorded sync.Map
}

func NewImageProxyContextHandler(p http.Handler) *ImageProxyContextHandler {
	return &ImageProxyContextHandler{proxy: p}
}

var (
	actorContextRe = regexp.MustCompile(`^act-([0-9]+)$`)
	iconContextRe  = regexp.MustCompile(`^icon-[a-z0-9]+$`)
)

func (h *ImageProxyContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/img/")
	seg := rest
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		seg = rest[:i]
	}
	if seg == "" {
		http.Error(w, "image proxy: missing context", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.recordAssociation(seg, remainder, r.URL.RawQuery)
	}

	http.StripPrefix("/img/"+seg, h.proxy).ServeHTTP(w, r)
}

// recordAssociation best-effort records the image<->entity association for
// this request. It must never affect serving the image.
func (h *ImageProxyContextHandler) recordAssociation(seg, remainder, rawQuery string) {
	if seg == "0" || seg == "act-0" {
		return
	}

	// remainder = "/<options>/<remote-url>"; options may be empty ("//...").
	// The remote url uses the ":/" scheme hack; its query string, if any,
	// arrived as this request's RawQuery.
	parts := strings.SplitN(strings.TrimPrefix(remainder, "/"), "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return
	}
	options, remoteURL := parts[0], strings.Replace(parts[1], ":/", "://", 1)
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

	if m := actorContextRe.FindStringSubmatch(seg); m != nil {
		if !rowExists(db.Model(&models.Actor{}).Where("id = ?", m[1])) {
			return
		}
	} else if !iconContextRe.MatchString(seg) {
		// scene context (soft-deleted scenes are excluded by gorm's scope)
		if !rowExists(db.Model(&models.Scene{}).Where("scene_id = ?", seg)) {
			return
		}
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

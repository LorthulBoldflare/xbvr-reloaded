package models

import (
	"crypto/md5" //nolint:gosec // dedupe hash only, not security-relevant (httpcache uses md5 the same way)
	"encoding/hex"
	"time"

	"github.com/jinzhu/gorm"
)

// ImageProxyEntry records which remote images were requested through the
// image proxy (/img/<context>/<options>/<url>) for which entity. Context is
// "scene-<scenes.scene_id>", "act-<actors.id>" or "icon-<slug>".
//
// The table is insert-only bookkeeping: rows are never deleted by the app
// (not even on scene/actor deletion), so the request handler's in-process
// dedupe map never goes stale.
type ImageProxyEntry struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Context string `gorm:"size:255;unique_index:uix_imageproxy_ctx_hash" json:"context"` // "scene-slr-1234", "act-45", "icon-baberotica"
	URLHash string `gorm:"size:32;unique_index:uix_imageproxy_ctx_hash" json:"url_hash"` // ImageProxyURLHash(url, options)
	URL     string `gorm:"size:1000" json:"url"`
	Options string `gorm:"size:100" json:"options"`
}

// ImageProxyURLHash derives the dedupe hash for a remote URL + options pair.
// Kept out of the URL column so the unique index stays small enough for MySQL.
func ImageProxyURLHash(url, options string) string {
	sum := md5.Sum([]byte(url + "\x1f" + options))
	return hex.EncodeToString(sum[:])
}

// RecordImageProxyEntry best-effort records an image<->entity association.
// Errors are logged at debug level and swallowed: recording must never affect
// serving the image.
func RecordImageProxyEntry(db *gorm.DB, context, url, options string) {
	if db == nil || context == "" || url == "" {
		return
	}

	entry := ImageProxyEntry{
		Context: context,
		URLHash: ImageProxyURLHash(url, options),
		URL:     url,
		Options: options,
	}
	err := db.Where(ImageProxyEntry{Context: context, URLHash: entry.URLHash}).
		Attrs(ImageProxyEntry{URL: url, Options: options}).
		FirstOrCreate(&entry).Error
	if err != nil {
		log.Debugf("imageproxy: failed to record entry for context %q: %v", context, err)
	}
}

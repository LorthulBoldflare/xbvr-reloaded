// getImageURL routes remote images through the image proxy at the given size
// (e.g. '700x', '120x', '700,fit', 'x40'). Local/relative URLs pass through
// unchanged; empty URLs return the optional blank placeholder.
//
// context identifies the entity the image belongs to and becomes the first
// path segment: a scene's scene_id (e.g. 'slr-1234', '0' if unknown),
// 'act-<actor id>', or 'icon-<slug>' (see iconSlug). Inserted raw, matching
// how scene ids are interpolated into /scenes/<scene_id> routes.
export function getImageURL (u, size = '700x', context = '0', blank = null) {
  if (u === '' || u === undefined || u === null) {
    return blank === null ? u : blank
  }
  if (!u.startsWith('http')) {
    return u
  }
  if (u.search('%') === -1) {
    return '/img/' + context + '/' + size + '/' + encodeURI(u)
  }
  // scraped URLs can contain malformed percent sequences (e.g. '100%.jpg')
  // that make decodeURI throw — fall back to the raw URL like the previous
  // per-component copies did
  try {
    return '/img/' + context + '/' + size + '/' + encodeURI(decodeURI(u))
  } catch (_) {
    return '/img/' + context + '/' + size + '/' + encodeURI(u)
  }
}

// iconSlug builds the slug for 'icon-<slug>' image proxy contexts: lowercased
// name with all special characters removed.
export function iconSlug (name) {
  return String(name).toLowerCase().replace(/[^a-z0-9]/g, '')
}

// altSourceIconContext builds the icon context for an AlternateSource's
// site_icon, deriving the slug from external_source (there is no plain site
// field on that DTO).
export function altSourceIconContext (altsrc) {
  return 'icon-' + iconSlug(String(altsrc.external_source).replace('alternate scene ', '').replace('stashdb scene', 'stashdb'))
}

export function humanizeSeconds (seconds) {
  return new Date(seconds * 1000).toISOString().substr(11, 8)
}

export function humanizeSeconds1DP (seconds) {
  return new Date(seconds * 1000).toISOString().substr(11, 10)
}

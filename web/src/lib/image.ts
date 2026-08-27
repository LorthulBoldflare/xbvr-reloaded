// Port of ui/src/util/image.js: remote image URLs are proxied+resized through
// the server's image proxy; local/relative URLs pass through unchanged.
//
// context identifies the entity the image belongs to and becomes the first
// path segment: a scene's scene_id (e.g. 'slr-1234', '0' if unknown),
// 'act-<actor id>', or 'icon-<slug>' (see iconSlug). Inserted raw, matching
// how scene ids are interpolated into /scenes/<scene_id> routes.

import type { AlternateSource } from '../api/types'

const BLANK = `${import.meta.env.BASE_URL}blank.png`

export function getImageURL(url: string | null | undefined, size: string, context: string): string {
  if (!url) return BLANK
  if (!url.startsWith('http')) return url
  if (url.search('%') === -1) {
    return `/img/${context}/${size}/${encodeURI(url)}`
  }
  // Scraped URLs can already be percent-encoded (or contain malformed
  // sequences like '100%.jpg' that make decodeURI throw). Decode first to
  // avoid double-encoding; fall back to the raw URL. MUST match the old UI's
  // util/image.js or proxied thumbnails 404.
  try {
    return `/img/${context}/${size}/${encodeURI(decodeURI(url))}`
  } catch {
    return `/img/${context}/${size}/${encodeURI(url)}`
  }
}

// iconSlug builds the slug for 'icon-<slug>' image proxy contexts: lowercased
// name with all special characters removed.
export function iconSlug(name: string): string {
  return String(name).toLowerCase().replace(/[^a-z0-9]/g, '')
}

// altSourceIconContext builds the icon context for an AlternateSource's
// site_icon, deriving the slug from external_source (there is no plain site
// field on that DTO). Mirrors AltSourcesSection's display name mangling.
export function altSourceIconContext(a: AlternateSource): string {
  return 'icon-' + iconSlug(a.external_source.replace('alternate scene ', '').replace('stashdb scene', 'stashdb'))
}

export function blankImage(): string {
  return BLANK
}

export function blankActorImage(): string {
  return BLANK
}

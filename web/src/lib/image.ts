// Port of ui/src/util/image.js: remote image URLs are proxied+resized through
// the server's image proxy; local/relative URLs pass through unchanged.

const BLANK = `${import.meta.env.BASE_URL}blank.png`

export function getImageURL(url: string | null | undefined, size = '700x'): string {
  if (!url) return BLANK
  if (!url.startsWith('http')) return url
  if (url.search('%') === -1) {
    return `/img/${size}/${encodeURI(url)}`
  }
  // Scraped URLs can already be percent-encoded (or contain malformed
  // sequences like '100%.jpg' that make decodeURI throw). Decode first to
  // avoid double-encoding; fall back to the raw URL. MUST match the old UI's
  // util/image.js or proxied thumbnails 404.
  try {
    return `/img/${size}/${encodeURI(decodeURI(url))}`
  } catch {
    return `/img/${size}/${encodeURI(url)}`
  }
}

export function blankImage(): string {
  return BLANK
}

export function blankActorImage(): string {
  return BLANK
}

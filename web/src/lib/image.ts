// Port of ui/src/util/image.js: remote image URLs are proxied+resized through
// the server's image proxy; local/relative URLs pass through unchanged.

const BLANK = `${import.meta.env.BASE_URL}blank.png`

export function getImageURL(url: string | null | undefined, size = '700x'): string {
  if (!url) return BLANK
  if (url.startsWith('http') || url.startsWith('https')) {
    return `/img/${size}/${encodeURI(url)}`
  }
  return url
}

export function blankImage(): string {
  return BLANK
}

export function blankActorImage(): string {
  return BLANK
}

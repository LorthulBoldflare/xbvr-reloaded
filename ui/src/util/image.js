// getImageURL routes remote images through the image proxy at the given size
// (e.g. '700x', '120x', '700,fit', 'x40'). Local/relative URLs pass through
// unchanged; empty URLs return the optional blank placeholder.
export function getImageURL (u, size = '700x', blank = null) {
  if (u === '' || u === undefined || u === null) {
    return blank === null ? u : blank
  }
  if (!u.startsWith('http')) {
    return u
  }
  if (u.search('%') === -1) {
    return '/img/' + size + '/' + encodeURI(u)
  }
  // scraped URLs can contain malformed percent sequences (e.g. '100%.jpg')
  // that make decodeURI throw — fall back to the raw URL like the previous
  // per-component copies did
  try {
    return '/img/' + size + '/' + encodeURI(decodeURI(u))
  } catch (_) {
    return '/img/' + size + '/' + encodeURI(u)
  }
}

export function humanizeSeconds (seconds) {
  return new Date(seconds * 1000).toISOString().substr(11, 8)
}

export function humanizeSeconds1DP (seconds) {
  return new Date(seconds * 1000).toISOString().substr(11, 10)
}

// Unicode-safe base64 helpers for encoding filter state into URLs.
// Replaces the Node `Buffer` polyfill usage.
export function encodeJsonBase64 (obj) {
  return btoa(unescape(encodeURIComponent(JSON.stringify(obj))))
}

export function decodeJsonBase64 (str) {
  return JSON.parse(decodeURIComponent(escape(atob(str))))
}

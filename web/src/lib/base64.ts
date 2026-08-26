// Unicode-safe base64 helpers for encoding filter state into URLs.
// MUST stay byte-compatible with ui/src/util/base64.js — saved searches and
// shared ?q= links are exchanged between the old and the new UI.
export function encodeJsonBase64(obj: unknown): string {
  return btoa(unescape(encodeURIComponent(JSON.stringify(obj))))
}

export function decodeJsonBase64<T>(str: string): T {
  return JSON.parse(decodeURIComponent(escape(atob(str)))) as T
}

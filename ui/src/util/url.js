// safeHref returns the URL only when it uses a safe scheme (http/https or
// site-relative); scraped content is attacker-influenced, so this blocks
// javascript: and other dangerous schemes in href targets.
export function safeHref (u) {
  if (typeof u !== 'string') {
    return '#'
  }
  const t = u.trim()
  if (t.startsWith('http://') || t.startsWith('https://') || t.startsWith('/')) {
    return t
  }
  return '#'
}

import prettyBytes from 'pretty-bytes'

export { prettyBytes }

// mm:ss, or h:mm:ss for >= 1h. Input is seconds.
export function humanizeSeconds(seconds: number): string {
  if (!seconds || seconds <= 0) return '0:00'
  const s = Math.floor(seconds % 60)
  const m = Math.floor((seconds / 60) % 60)
  const h = Math.floor(seconds / 3600)
  const pad = (n: number) => String(n).padStart(2, '0')
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`
}

// seconds → "h:mm:ss.s"
export function humanizeSeconds1DP(seconds: number): string {
  if (!seconds || seconds <= 0) return '0:00.0'
  const base = humanizeSeconds(Math.floor(seconds))
  return `${base}.${Math.floor((seconds % 1) * 10)}`
}

const pad2 = (n: number) => String(n).padStart(2, '0')

// Date-only strings pass through verbatim; full timestamps (RFC3339 from the
// server) are rendered in local time, matching the old date-fns behavior —
// taking the literal prefix would show the UTC date for users behind UTC.
export function formatDate(iso: string | null | undefined): string {
  if (!iso || iso.startsWith('0001-01-01')) return ''
  const dateOnly = /^(\d{4}-\d{2}-\d{2})$/.exec(iso)
  if (dateOnly) return dateOnly[1]
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())}`
}

export function formatDateTime(iso: string | null | undefined): string {
  if (!iso || iso.startsWith('0001-01-01')) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}`
}

// Only http(s) and site-relative URLs may be used as link targets (port of
// ui/src/util/url.js safeHref).
export function safeHref(url: string | null | undefined): string {
  if (!url) return '#'
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('/')) return url
  return '#'
}

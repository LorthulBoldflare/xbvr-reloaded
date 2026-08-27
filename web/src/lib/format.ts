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

export function formatDate(iso: string | null | undefined): string {
  if (!iso || iso.startsWith('0001-01-01')) return ''
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(iso)
  return match ? `${match[1]}-${match[2]}-${match[3]}` : ''
}

export function formatDateTime(iso: string | null | undefined): string {
  if (!iso || iso.startsWith('0001-01-01')) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

// Only http(s) and site-relative URLs may be used as link targets (port of
// ui/src/util/url.js safeHref).
export function safeHref(url: string | null | undefined): string {
  if (!url) return '#'
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('/')) return url
  return '#'
}

// API client for the XBVR server. All paths are relative to /api.
// Auth is handled by the browser (Basic auth credentials are cached per-origin,
// or the player-session cookie is sent automatically) — no client-side auth code.

import { useToastStore } from '../store/toasts'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

type Options = {
  json?: unknown
  // AbortSignal for the request.
  signal?: AbortSignal
  // Set false to suppress the global error toast.
  toastOnError?: boolean
}

async function request<T>(method: string, path: string, opts: Options = {}): Promise<T> {
  const url = path.startsWith('http') ? path : `/api${path}`
  const init: RequestInit = {
    method,
    headers: opts.json !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: opts.json !== undefined ? JSON.stringify(opts.json) : undefined,
    signal: opts.signal
  }
  let res: Response
  try {
    res = await fetch(url, init)
  } catch (e) {
    if ((e as Error).name === 'AbortError') throw e
    if (opts.toastOnError !== false) useToastStore.getState().error(`Request failed: ${method} ${path}`)
    throw e
  }
  if (!res.ok) {
    if (res.status === 401 && opts.toastOnError !== false) {
      useToastStore.getState().error('Authentication required')
    } else if (opts.toastOnError !== false) {
      const body = await res.text().catch(() => '')
      useToastStore.getState().error(`Request failed: ${method} ${path} — ${res.status} ${body.slice(0, 200)}`)
    }
    throw new ApiError(res.status, `${method} ${path} failed: ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  const text = await res.text()
  if (!text) return undefined as T
  return JSON.parse(text) as T
}

export const api = {
  get: <T>(path: string, opts?: Options) => request<T>('GET', path, opts),
  post: <T>(path: string, json?: unknown, opts?: Options) => request<T>('POST', path, { ...opts, json }),
  put: <T>(path: string, json?: unknown, opts?: Options) => request<T>('PUT', path, { ...opts, json }),
  // The Go API expects the body for several DELETE endpoints (e.g. extref) —
  // keep parity with the old UI by allowing a JSON body on DELETE.
  delete: <T>(path: string, json?: unknown, opts?: Options) => request<T>('DELETE', path, { ...opts, json })
}

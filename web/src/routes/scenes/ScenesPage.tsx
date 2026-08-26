import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { File, ResponseSceneList, Scene, SceneFilters } from '../../api/types'
import {
  applyDlState,
  DEFAULT_SCENE_FILTERS,
  sceneListRequestBody,
  useSceneFilterStore
} from '../../store/sceneFilters'
import { decodeJsonBase64, encodeJsonBase64 } from '../../lib/base64'
import { SCENE_SORTS } from '../../api/sorts'
import { SceneCard } from '../../components/SceneCard'
import { FileCard } from '../../components/FileCard'
import { Popover } from '../../components/Popover'
import { FiltersPopoverContent } from './FiltersPopover'

const PAGE_SIZE = 80

function newestVideoTime(s: Scene): number {
  let max = 0
  for (const f of s.file ?? []) {
    if (f.type === 'video') {
      const t = Date.parse(f.created_time)
      if (t > max) max = t
    }
  }
  return max
}

// The scene grid is the app's main page. Downloaded-first; the Browse tab
// shows the whole library (incl. not-downloaded) as a denser grid.
export function ScenesPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const { filters, setFilters, patch } = useSceneFilterStore()
  const [autoLoad, setAutoLoad] = useState(true)

  // ---- URL → store (on mount / external nav), store → URL (on change)
  const lastQ = useRef<string | null>(null)
  const didInit = useRef(false)
  useEffect(() => {
    const q = searchParams.get('q')
    try {
      if (q !== lastQ.current) {
        lastQ.current = q
        if (q) {
          const parsed = decodeJsonBase64<Partial<SceneFilters>>(q)
          // Tolerant parse: tolerate old-UI-only keys (cardSize etc.)
          const merged = { ...DEFAULT_SCENE_FILTERS, ...parsed }
          setFilters(applyDlState(merged, merged.dlState ?? 'available'))
        } else {
          setFilters(DEFAULT_SCENE_FILTERS)
        }
      }
    } finally {
      // Never let the store→URL effect write defaults over an unparsed
      // incoming deep link.
      didInit.current = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams])

  useEffect(() => {
    if (!didInit.current) return
    // Read the store imperatively: in the commit where the URL→store effect
    // just applied an incoming deep link, the render-scoped `filters` closure
    // is still stale (defaults) and would clobber the URL.
    const encoded = encodeJsonBase64(useSceneFilterStore.getState().filters)
    if (encoded !== searchParams.get('q')) {
      lastQ.current = encoded
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          next.set('q', encoded)
          return next
        },
        { replace: true }
      )
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters])

  // Legacy deep link: /?scene_id=... → scene page
  useEffect(() => {
    const sid = searchParams.get('scene_id')
    if (sid) navigate(`/scenes/${sid}`, { replace: true })
  }, [searchParams, navigate])

  // ---- Data
  const listQuery = useInfiniteQuery({
    queryKey: ['sceneList', filters],
    queryFn: ({ pageParam }) =>
      api.post<ResponseSceneList>('/scene/list', sceneListRequestBody(filters, pageParam as number, PAGE_SIZE)),
    initialPageParam: 0,
    getNextPageParam: (lastPage, pages) => {
      const loaded = pages.length * PAGE_SIZE
      return loaded < lastPage.results ? loaded : undefined
    }
  })

  const scenes: Scene[] = useMemo(() => listQuery.data?.pages.flatMap((p) => p.scenes) ?? [], [listQuery.data])
  const firstPage = listQuery.data?.pages[0]

  const unmatchedQuery = useQuery({
    queryKey: ['unmatchedFiles'],
    queryFn: () => api.post<File[]>('/files/list', { state: 'unmatched', sort: 'created_time_desc' }),
    enabled: !!filters.showUnmatched,
    staleTime: 30_000
  })

  // Merge unmatched files into the grid (see plan: by created_time for the
  // file_added sorts, appended otherwise).
  const items = useMemo(() => {
    const sceneItems = scenes.map((s) => ({ kind: 'scene' as const, scene: s }))
    if (!filters.showUnmatched) return sceneItems
    const fileItems = (unmatchedQuery.data ?? []).map((f) => ({ kind: 'file' as const, file: f }))
    if (fileItems.length === 0) return sceneItems
    if (filters.sort.startsWith('file_added')) {
      const combined = [
        ...sceneItems.map((i) => ({ t: newestVideoTime(i.scene), i })),
        ...fileItems.map((i) => ({ t: Date.parse(i.file.created_time), i }))
      ]
      combined.sort((a, b) => (filters.sort === 'file_added_asc' ? a.t - b.t : b.t - a.t))
      return combined.map((c) => c.i)
    }
    return [...sceneItems, ...fileItems]
  }, [scenes, unmatchedQuery.data, filters.showUnmatched, filters.sort])

  // ---- Infinite scroll
  const sentinelRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!autoLoad) return
    const el = sentinelRef.current
    if (!el) return
    const obs = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting && listQuery.hasNextPage && !listQuery.isFetchingNextPage) {
        listQuery.fetchNextPage()
      }
    })
    obs.observe(el)
    return () => obs.disconnect()
  }, [autoLoad, listQuery])

  // Keep the loaded order for scene-page prev/next navigation.
  useEffect(() => {
    sessionStorage.setItem('sceneOrder', JSON.stringify(scenes.map((s) => s.scene_id)))
  }, [scenes])

  // Active-filter chips for the top bar
  const chips = useMemo(() => {
    const out: { key: string; label: string; clear: () => void }[] = []
    if (filters.dlState !== 'available') {
      out.push({
        key: 'dl',
        label: `availability: ${filters.dlState === 'missing' ? 'not downloaded' : filters.dlState}`,
        clear: () => setFilters(applyDlState(filters, 'available'))
      })
    }
    for (const l of filters.lists) out.push({ key: `l:${l}`, label: l, clear: () => patch({ lists: filters.lists.filter((x) => x !== l) }) })
    const strip = (v: string) => (v.startsWith('!') || v.startsWith('&') ? v.slice(1) : v)
    const mk = (field: 'cast' | 'sites' | 'tags' | 'cuepoint' | 'attributes', prefix: string) =>
      filters[field].forEach((v) =>
        out.push({
          key: `${field}:${v}`,
          label: `${prefix}: ${v}`,
          clear: () => patch({ [field]: filters[field].filter((x) => x !== v) } as never)
        })
      )
    mk('cast', 'cast')
    mk('sites', 'site')
    mk('tags', 'tag')
    mk('cuepoint', 'cuepoint')
    mk('attributes', 'attr')
    if (filters.isWatched !== null)
      out.push({ key: 'w', label: filters.isWatched ? 'watched' : 'unwatched', clear: () => patch({ isWatched: null }) })
    if (filters.releaseMonth)
      out.push({ key: 'rm', label: `released: ${filters.releaseMonth}`, clear: () => patch({ releaseMonth: '' }) })
    if (filters.volume) out.push({ key: 'vol', label: 'folder selected', clear: () => patch({ volume: 0 }) })
    return out
  }, [filters, patch, setFilters])

  const counts = firstPage
    ? {
        count_any: firstPage.count_any,
        count_available: firstPage.count_available,
        count_downloaded: firstPage.count_downloaded,
        count_not_downloaded: firstPage.count_not_downloaded,
        count_hidden: firstPage.count_hidden
      }
    : undefined

  return (
    <div>
      {/* Top bar: filters popover + sort + chips + count */}
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <Popover
          button={
            <>
              Filters
              {chips.length > 0 && (
                <span className="rounded-full bg-accent px-1.5 text-[10px] font-bold text-white">{chips.length}</span>
              )}
              <span className="text-muted">▾</span>
            </>
          }
          width="w-[26rem]"
        >
          <FiltersPopoverContent counts={counts} />
        </Popover>

        <select
          value={filters.sort}
          onChange={(e) => patch({ sort: e.target.value })}
          className="rounded-lg border border-line bg-surface px-2.5 py-1.5 text-sm"
          title="Sort"
        >
          {SCENE_SORTS.map((s) => (
            <option key={s.value} value={s.value}>
              {s.label}
            </option>
          ))}
        </select>

        <div className="flex min-w-0 flex-1 flex-wrap items-center gap-1">
          {chips.map((c) => (
            <span
              key={c.key}
              className="flex items-center gap-1 rounded-full bg-surface-3 px-2 py-0.5 text-xs text-fg"
            >
              {c.label}
              <button onClick={c.clear} className="text-muted hover:text-danger" aria-label="Remove filter">
                ✕
              </button>
            </span>
          ))}
        </div>

        <span className="text-sm text-muted">{firstPage ? `${firstPage.results} results` : ''}</span>
        <button
          onClick={() => setAutoLoad((a) => !a)}
          title="Toggle auto load more"
          className={`rounded-lg border px-2 py-1.5 text-xs ${autoLoad ? 'border-accent text-accent-strong' : 'border-line text-muted'}`}
        >
          auto-load
        </button>
      </div>

      {/* Grid */}
      {listQuery.isLoading && <div className="py-16 text-center text-muted">Loading…</div>}
      {listQuery.isError && <div className="py-16 text-center text-danger">Failed to load scenes</div>}
      <div className="grid grid-cols-[repeat(auto-fill,minmax(230px,1fr))] gap-4">
        {items.map((item) =>
          item.kind === 'scene' ? (
            <SceneCard key={`s${item.scene.id}`} scene={item.scene} />
          ) : (
            <FileCard key={`f${item.file.id}`} file={item.file} />
          )
        )}
      </div>
      {!listQuery.isLoading && items.length === 0 && (
        <div className="py-16 text-center text-muted">No scenes match the current filters</div>
      )}

      <div ref={sentinelRef} />
      {listQuery.isFetchingNextPage && <div className="py-4 text-center text-muted">Loading more…</div>}
      {!autoLoad && listQuery.hasNextPage && (
        <div className="py-4 text-center">
          <button
            onClick={() => listQuery.fetchNextPage()}
            className="rounded-lg border border-line bg-surface px-4 py-2 text-sm hover:bg-surface-2"
          >
            Load more
          </button>
        </div>
      )}
    </div>
  )
}

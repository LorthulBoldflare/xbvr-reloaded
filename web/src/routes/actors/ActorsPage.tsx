import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { Actor, ActorFilters, ResponseActorList } from '../../api/types'
import { ACTOR_PAGE_SIZE, ACTOR_SORTS, DEFAULT_ACTOR_FILTERS, useActorFilterStore } from '../../store/actorFilters'
import { decodeJsonBase64, encodeJsonBase64 } from '../../lib/base64'
import { useUIStore } from '../../store/ui'
import { ActorCard } from '../../components/ActorCard'
import { Popover } from '../../components/Popover'
import { ActorFiltersPopoverContent } from './ActorFiltersPopover'

const LETTERS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('')

// Actors list: paginated (not infinite scroll), A–Z jump bar on name sorts.
export function ActorsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { filters, setFilters, patch } = useActorFilterStore()
  const [page, setPage] = useState(0)
  const queryClient = useQueryClient()
  const showActorDetails = useUIStore((s) => s.showActorDetails)
  const actorDetailsOpen = useUIStore((s) => s.actorDetailsId !== null)
  const editActorOpen = useUIStore((s) => s.editActorId !== null)

  // URL sync (same ?q= contract as scenes)
  const didInit = useRef(false)
  useEffect(() => {
    const q = searchParams.get('q')
    if (q) {
      try {
        setFilters({ ...DEFAULT_ACTOR_FILTERS, ...decodeJsonBase64<Partial<ActorFilters>>(q) })
      } catch {
        /* ignore */
      }
    }
    const actorId = searchParams.get('actor_id')
    if (actorId) showActorDetails(Number(actorId))
    didInit.current = true
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!didInit.current) return
    const encoded = encodeJsonBase64(filters)
    if (encoded !== searchParams.get('q')) {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          next.set('q', encoded)
          return next
        },
        { replace: true }
      )
    }
    setPage(0)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters])

  const { data } = useQuery({
    queryKey: ['actorList', filters, page],
    queryFn: () =>
      api.post<ResponseActorList>('/actor/list', { ...filters, offset: page * ACTOR_PAGE_SIZE, limit: ACTOR_PAGE_SIZE }),
    placeholderData: (prev) => prev
  })

  const total = data?.results ?? 0
  const pageCount = Math.max(1, Math.ceil(total / ACTOR_PAGE_SIZE))
  const actors = data?.actors ?? []

  // Keep the loaded order for actor-modal prev/next navigation.
  useEffect(() => {
    sessionStorage.setItem('actorOrder', JSON.stringify(actors.map((a) => a.id)))
  }, [actors])

  // A–Z jump: only meaningful for name sorts
  const showJump = filters.sort === 'name_asc' || filters.sort === 'name_desc' || filters.sort === ''
  const jump = (letter: string) => {
    patch({ jumpTo: letter })
    // server returns the new offset for the letter
  }
  useEffect(() => {
    if (data && filters.jumpTo) {
      setPage(Math.floor((data.offset ?? 0) / ACTOR_PAGE_SIZE))
      patch({ jumpTo: '' })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  // keyboard paging (←/→, o/p), suppressed while an overlay is open
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      if (['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) || target.isContentEditable) return
      if (actorDetailsOpen || editActorOpen) return
      if (e.key === 'ArrowRight' || e.key === 'p') setPage((p) => (p + 1) % pageCount)
      if (e.key === 'ArrowLeft' || e.key === 'o') setPage((p) => (p - 1 + pageCount) % pageCount)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [pageCount, actorDetailsOpen, editActorOpen])

  const pager = (
    <div className="flex items-center gap-2 text-sm">
      <button
        disabled={page === 0}
        onClick={() => setPage((p) => p - 1)}
        className="rounded-lg border border-line px-2 py-1 disabled:opacity-40"
      >
        ←
      </button>
      <span className="text-muted">
        Page{' '}
        <input
          value={page + 1}
          onChange={(e) => {
            const n = Number(e.target.value)
            if (n >= 1 && n <= pageCount) setPage(n - 1)
          }}
          className="w-12 rounded border border-line bg-surface-2 px-1 py-0.5 text-center"
        />{' '}
        / {pageCount}
      </span>
      <button
        disabled={page >= pageCount - 1}
        onClick={() => setPage((p) => p + 1)}
        className="rounded-lg border border-line px-2 py-1 disabled:opacity-40"
      >
        →
      </button>
    </div>
  )

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <Popover button={<>Filters <span className="text-muted">▾</span></>} width="w-[26rem]">
          <ActorFiltersPopoverContent />
        </Popover>
        <select
          value={filters.sort}
          onChange={(e) => patch({ sort: e.target.value })}
          className="rounded-lg border border-line bg-surface px-2.5 py-1.5 text-sm"
          title="Sort"
        >
          {ACTOR_SORTS.map((s) => (
            <option key={s.value} value={s.value}>
              {s.label}
            </option>
          ))}
        </select>
        <span className="flex-1" />
        <span className="text-sm text-muted">{total} results</span>
        {pager}
      </div>

      {showJump && (
        <div className="mb-3 flex flex-wrap gap-0.5">
          {LETTERS.map((l) => (
            <button
              key={l}
              onClick={() => jump(l)}
              className="rounded px-1.5 py-0.5 text-xs font-semibold text-muted hover:bg-accent-soft hover:text-accent-strong"
            >
              {l}
            </button>
          ))}
        </div>
      )}

      <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-4">
        {actors.map((a) => (
          <ActorCard key={a.id} actor={a} />
        ))}
      </div>
      {actors.length === 0 && <div className="py-16 text-center text-muted">No actors match the current filters</div>}

      <div className="mt-4 flex justify-center">{pager}</div>
    </div>
  )
}

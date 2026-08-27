import { useEffect, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Actor, StashdbPerformerSearchResponse } from '../api/types'
import { useUIStore } from '../store/ui'
import { useToastStore } from '../store/toasts'
import { getImageURL } from '../lib/image'
import { Modal } from './Modal'
import { LinkIcon } from './icons'

// StashDB performer search + link for actors.
export function SearchStashdbActors() {
  const actorId = useUIStore((s) => s.stashdbActorSearchId)
  const hide = useUIStore((s) => s.hideStashdbActorSearch)
  const queryClient = useQueryClient()
  const toast = useToastStore.getState()

  const { data: actor } = useQuery({
    queryKey: ['actor', actorId],
    queryFn: () => api.get<Actor>(`/actor/${actorId}`),
    enabled: actorId !== null
  })

  const [query, setQuery] = useState('')
  const [debounced, setDebounced] = useState('')
  const [imgIdx, setImgIdx] = useState<Record<string, number>>({})

  useEffect(() => {
    if (actorId !== null && actor) {
      setQuery(actor.name)
      setDebounced(actor.name)
    }
  }, [actorId, actor])

  useEffect(() => {
    const t = setTimeout(() => setDebounced(query), 750)
    return () => clearTimeout(t)
  }, [query])

  const { data, isFetching } = useQuery({
    queryKey: ['stashdbActorSearch', actorId, debounced],
    queryFn: ({ signal }) =>
      api.get<StashdbPerformerSearchResponse>(
        `/extref/stashdb/searchactor/${actorId}?q=${encodeURIComponent(debounced)}`,
        { signal, toastOnError: false }
      ),
    enabled: actorId !== null
  })

  useEffect(() => {
    if (data?.Status) toast.error(data.Status)
  }, [data?.Status, toast])

  const results = [...(data?.Results ?? [])].sort((a, b) => b.Weight - a.Weight)
  const aliasSet = new Set(
    [actor?.name ?? '', ...(() => {
      try {
        return JSON.parse(actor?.aliases || '[]') as string[]
      } catch {
        return []
      }
    })()].map((s) => s.toLowerCase())
  )

  const link = async (stashUrl: string) => {
    const stashId = stashUrl.replace('https://stashdb.org/performers/', '')
    try {
      await api.get(`/extref/stashdb/link2actor/${actorId}/${stashId}`)
      toast.success('Linked to StashDB')
      queryClient.invalidateQueries({ queryKey: ['actor', actorId] })
      queryClient.invalidateQueries({ queryKey: ['actorExtrefs', actorId] })
      hide()
    } catch {
      /* toast shown by client */
    }
  }

  return (
    <Modal open={actorId !== null} onClose={hide} width="max-w-4xl" title="Search StashDB performers">
      <input
        autoFocus
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="mb-3 w-full rounded-lg border border-line bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
      />
      {isFetching && <div className="py-2 text-xs text-muted">Searching…</div>}
      <div className="max-h-[60vh] space-y-2 overflow-y-auto">
        {results.map((r) => {
          const idx = imgIdx[r.Id] ?? 0
          const imgs = r.ImageUrl ?? []
          return (
            <div key={r.Id} className="flex gap-3 rounded-lg border border-line bg-surface-2 p-2">
              <div className="relative h-24 w-24 shrink-0">
                {imgs.length > 0 && (
                  <img src={getImageURL(imgs[Math.min(idx, imgs.length - 1)], '120x', 'act-0')} alt="" className="h-24 w-24 rounded object-cover" loading="lazy" />
                )}
                {imgs.length > 1 && (
                  <div className="absolute inset-x-0 bottom-0 flex justify-center gap-1">
                    <button
                      onClick={() => setImgIdx((m) => ({ ...m, [r.Id]: (idx - 1 + imgs.length) % imgs.length }))}
                      className="rounded bg-black/60 px-1 text-xs text-white"
                    >
                      ←
                    </button>
                    <button
                      onClick={() => setImgIdx((m) => ({ ...m, [r.Id]: (idx + 1) % imgs.length }))}
                      className="rounded bg-black/60 px-1 text-xs text-white"
                    >
                      →
                    </button>
                  </div>
                )}
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-xs text-muted">
                  {r.DOB && <span>{r.DOB}</span>}
                  <span className="rounded bg-surface-3 px-1.5 font-semibold">{Math.round(r.Weight)}</span>
                  <a href={r.Url} target="_blank" rel="noreferrer" className="hover:text-accent">
                    <LinkIcon />
                  </a>
                </div>
                <a href={r.Url} target="_blank" rel="noreferrer" className="block truncate text-sm font-semibold hover:text-accent">
                  {r.Name}
                  {r.Disambiguation && <span className="font-normal text-muted"> — {r.Disambiguation}</span>}
                </a>
                <div className="mt-0.5 flex flex-wrap gap-1">
                  {(r.Aliases ?? []).map((a) => (
                    <span
                      key={a}
                      className={`rounded-full px-1.5 py-0.5 text-[10px] ${aliasSet.has(a.toLowerCase()) ? 'bg-ok/20 font-bold text-ok' : 'bg-surface-3 text-muted'}`}
                    >
                      {a}
                    </span>
                  ))}
                </div>
                <div className="mt-0.5 flex flex-wrap gap-1">
                  {(r.Studios ?? []).map((s) => (
                    <span key={s.Name} className="rounded-full bg-surface-3 px-1.5 py-0.5 text-[10px] text-muted">
                      {s.Name} ({s.SceneCount})
                    </span>
                  ))}
                </div>
              </div>
              <button
                onClick={() => link(r.Url)}
                className="self-center rounded-lg bg-accent px-3 py-1.5 text-xs font-semibold text-white"
              >
                Link
              </button>
            </div>
          )
        })}
        {data && results.length === 0 && <div className="py-6 text-center text-sm text-muted">No results</div>}
      </div>
    </Modal>
  )
}

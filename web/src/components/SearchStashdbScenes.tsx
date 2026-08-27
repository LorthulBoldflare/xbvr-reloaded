import { useEffect, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Scene, StashdbSceneSearchResponse } from '../api/types'
import { useUIStore } from '../store/ui'
import { useToastStore } from '../store/toasts'
import { getImageURL, sceneContext } from '../lib/image'
import { humanizeSeconds } from '../lib/format'
import { Modal } from './Modal'
import { LinkIcon } from './icons'

// StashDB scene search + link (launched from the scene page edit mode).
export function SearchStashdbScenes() {
  const sceneId = useUIStore((s) => s.stashdbSceneSearchId)
  const hide = useUIStore((s) => s.hideStashdbSceneSearch)
  const queryClient = useQueryClient()
  const toast = useToastStore()

  const [query, setQuery] = useState('')
  const [debounced, setDebounced] = useState('')

  useEffect(() => {
    if (sceneId !== null) {
      setQuery('')
      setDebounced('')
    }
  }, [sceneId])

  useEffect(() => {
    const t = setTimeout(() => setDebounced(query), 750)
    return () => clearTimeout(t)
  }, [query])

  const { data, isFetching } = useQuery({
    queryKey: ['stashdbSceneSearch', sceneId, debounced],
    queryFn: () =>
      api.get<StashdbSceneSearchResponse>(
        `/extref/stashdb/search/${sceneId}?q=${encodeURIComponent(debounced)}`,
        { toastOnError: false }
      ),
    enabled: sceneId !== null
  })

  useEffect(() => {
    if (data?.Status) toast.error(data.Status)
  }, [data?.Status, toast])

  const results = [...(data?.Results ?? [])].sort((a, b) => b.Weight - a.Weight)

  const link = async (stashUrl: string) => {
    const stashId = stashUrl.replace('https://stashdb.org/scenes/', '')
    try {
      const scene = await api.get<Scene>(`/extref/stashdb/link2scene/${sceneId}/${stashId}`)
      toast.success('Linked to StashDB')
      queryClient.setQueryData(['scene', 'sid', scene.scene_id], scene)
      queryClient.invalidateQueries({ queryKey: ['altSources', scene.id] })
      hide()
    } catch {
      /* toast already shown by client */
    }
  }

  return (
    <Modal open={sceneId !== null} onClose={hide} width="max-w-4xl" title="Search StashDB">
      <input
        autoFocus
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search StashDB…"
        className="mb-3 w-full rounded-lg border border-line bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
      />
      {isFetching && <div className="py-2 text-xs text-muted">Searching…</div>}
      <div className="max-h-[60vh] space-y-2 overflow-y-auto">
        {results.map((r) => (
          <div key={r.Id} className="flex gap-3 rounded-lg border border-line bg-surface-2 p-2">
            <img
              src={getImageURL(r.ImageUrl, '120x', sceneContext('stash-' + r.Id))}
              alt=""
              loading="lazy"
              className="h-24 w-24 shrink-0 rounded object-cover"
            />
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-xs text-muted">
                <span>{r.Date}</span>
                {r.Duration > 0 && <span>{humanizeSeconds(r.Duration)}</span>}
                <span className="rounded bg-surface-3 px-1.5 font-semibold">{Math.round(r.Weight)}</span>
                <a href={r.Url} target="_blank" rel="noreferrer" className="hover:text-accent" title="Open on StashDB">
                  <LinkIcon />
                </a>
              </div>
              <a href={r.Url} target="_blank" rel="noreferrer" className="block truncate text-sm font-semibold hover:text-accent">
                {r.Studio} — {r.Title}
              </a>
              <div className="truncate text-xs text-muted">{r.Performers.join(', ')}</div>
              {r.Description && <div className="mt-0.5 line-clamp-2 text-xs text-muted">{r.Description}</div>}
            </div>
            <button
              onClick={() => link(r.Url)}
              className="self-center rounded-lg bg-accent px-3 py-1.5 text-xs font-semibold text-white"
            >
              Link
            </button>
          </div>
        ))}
        {data && results.length === 0 && <div className="py-6 text-center text-sm text-muted">No results</div>}
      </div>
    </Modal>
  )
}

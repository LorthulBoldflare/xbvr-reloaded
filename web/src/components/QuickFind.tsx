import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { ResponseGetScenes, Scene } from '../api/types'
import { useUIStore } from '../store/ui'
import { getImageURL, sceneContext } from '../lib/image'
import { formatDate } from '../lib/format'
import { Modal } from './Modal'
import { StarRating } from './StarRating'

const HINTS = ['+title:', 'cast:', '+site:', '+id:', 'duration:>=0', 'released:>="', 'added:>="']

// Command-palette scene search (opened with "?"). In select mode
// (quickFindForSelect) the choice is stored in the UI store instead of
// navigating — used by the scene-page relink flow.
export function QuickFind() {
  const open = useUIStore((s) => s.quickFindOpen)
  const forSelect = useUIStore((s) => s.quickFindForSelect)
  const closeQuickFind = useUIStore((s) => s.closeQuickFind)
  const setSelected = useUIStore((s) => s.setQuickFindSelectedScene)
  const navigate = useNavigate()

  const [query, setQuery] = useState('')
  const [debounced, setDebounced] = useState('')
  const [active, setActive] = useState(0)
  const listRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (open) {
      setQuery('')
      setDebounced('')
      setActive(0)
    }
  }, [open])

  useEffect(() => {
    const t = setTimeout(() => setDebounced(query), 250)
    return () => clearTimeout(t)
  }, [query])

  const { data } = useQuery({
    queryKey: ['quickfind', debounced],
    queryFn: () => api.get<ResponseGetScenes>(`/scene/search?q=${encodeURIComponent(debounced)}`, { toastOnError: false }),
    enabled: open && debounced.length > 0,
    placeholderData: (prev) => prev
  })
  const results: Scene[] = data?.scenes ?? []

  const pick = (scene: Scene) => {
    if (forSelect) {
      setSelected(scene)
    } else {
      closeQuickFind()
      navigate(`/scenes/${scene.scene_id}`)
    }
  }

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActive((a) => Math.min(a + 1, results.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActive((a) => Math.max(a - 1, 0))
    } else if (e.key === 'Enter' && results[active]) {
      e.preventDefault()
      pick(results[active])
    }
  }

  useEffect(() => {
    listRef.current?.querySelector(`[data-idx="${active}"]`)?.scrollIntoView({ block: 'nearest' })
  }, [active])

  return (
    <Modal open={open} onClose={closeQuickFind} width="max-w-2xl">
      <input
        autoFocus
        value={query}
        onChange={(e) => {
          setQuery(e.target.value)
          setActive(0)
        }}
        onKeyDown={onKeyDown}
        placeholder={forSelect ? 'Search a scene to select…' : 'Search scenes…'}
        className="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
      />
      <div className="mt-2 flex flex-wrap gap-1">
        {HINTS.map((h) => (
          <button
            key={h}
            onClick={() => setQuery((q) => (q ? `${q} ${h}` : h))}
            className="rounded border border-line bg-surface-2 px-1.5 py-0.5 font-mono text-[11px] text-muted hover:border-accent hover:text-fg"
          >
            {h}
          </button>
        ))}
      </div>
      <div ref={listRef} className="mt-3 max-h-[50vh] overflow-y-auto">
        {results.map((scene, idx) => (
          <button
            key={scene.id}
            data-idx={idx}
            onClick={() => pick(scene)}
            onMouseEnter={() => setActive(idx)}
            className={`flex w-full items-center gap-3 rounded-lg px-2 py-1.5 text-left ${
              idx === active ? 'bg-accent-soft' : ''
            }`}
          >
            <img
              src={getImageURL(scene.cover_url, '120x', sceneContext(scene.scene_id))}
              alt=""
              className="h-12 w-12 shrink-0 rounded object-cover"
              loading="lazy"
            />
            <span className="min-w-0 flex-1">
              <span className="flex items-center gap-2 text-xs text-muted">
                <span>{scene.site}</span>
                {scene.is_hidden && (
                  <span className="rounded bg-surface-3 px-1 text-[10px] uppercase text-muted">hidden</span>
                )}
                <StarRating value={scene.star_rating} readonly size="sm" />
                <span>{formatDate(scene.release_date)}</span>
              </span>
              <span className="block truncate text-sm font-medium">{scene.title}</span>
              <span className="block truncate text-xs text-muted">
                {(scene.cast ?? []).map((a) => a.name).join(', ')}
              </span>
            </span>
          </button>
        ))}
        {debounced.length > 0 && results.length === 0 && (
          <div className="py-6 text-center text-sm text-muted">No results</div>
        )}
      </div>
      <div className="mt-3 border-t border-line pt-2 text-xs text-muted">
        ↑↓ navigate · ↵ open · esc close
      </div>
    </Modal>
  )
}

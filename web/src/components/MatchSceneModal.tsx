import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { File, ResponseGetScenes, Scene } from '../api/types'
import { getImageURL, sceneContext } from '../lib/image'
import { formatDate } from '../lib/format'
import { cleanFilename } from '../lib/filename'
import { Modal } from './Modal'

const PREFIX_CHIPS = ['+title:', 'cast:', '+site:', '+id:']

// Match an (unmatched) file to a scene. Launched from the file page or the
// Options → Files table.
export function MatchSceneModal({
  file,
  open,
  onClose,
  onMatched
}: {
  file: File | null
  open: boolean
  onClose: () => void
  onMatched: () => void
}) {
  const [query, setQuery] = useState('')
  const [debounced, setDebounced] = useState('')

  useEffect(() => {
    if (open && file) {
      const q = cleanFilename(file)
      setQuery(q)
      setDebounced(q)
    }
  }, [open, file])

  useEffect(() => {
    const t = setTimeout(() => setDebounced(query), 200)
    return () => clearTimeout(t)
  }, [query])

  const { data, isFetching } = useQuery({
    queryKey: ['sceneMatch', debounced],
    queryFn: ({ signal }) =>
      api.get<ResponseGetScenes>(`/scene/search?q=${encodeURIComponent(debounced)}`, {
        signal,
        toastOnError: false
      }),
    enabled: open && debounced.length > 0,
    placeholderData: (prev) => prev
  })
  const results = data?.scenes ?? []
  const maxScore = Math.max(1, ...results.map((r) => r._score ?? 0))

  const assign = async (scene: Scene) => {
    if (!file) return
    await api.post('/files/match', { file_id: file.id, scene_id: scene.scene_id })
    onMatched()
    onClose()
  }

  const addChip = (chip: string) => {
    if (chip === 'duration:' && file?.duration) {
      const mins = Math.round(file.duration / 60)
      setQuery((q) => `${q} duration:>=${mins - 1} duration:<=${mins + 1}`.trim())
      return
    }
    setQuery((q) => `${q} ${chip}`.trim())
  }

  return (
    <Modal open={open} onClose={onClose} width="max-w-3xl" title={`Match file to scene`}>
      {file && <div className="mb-2 truncate font-mono text-xs text-muted">{file.filename}</div>}
      <input
        autoFocus
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="mb-2 w-full rounded-lg border border-line bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
      />
      <div className="mb-3 flex flex-wrap gap-1">
        {PREFIX_CHIPS.map((c) => (
          <button
            key={c}
            onClick={() => addChip(c)}
            className="rounded border border-line bg-surface-2 px-1.5 py-0.5 font-mono text-[11px] text-muted hover:border-accent hover:text-fg"
          >
            {c}
          </button>
        ))}
        <button
          onClick={() => addChip('duration:')}
          className="rounded border border-line bg-surface-2 px-1.5 py-0.5 font-mono text-[11px] text-muted hover:border-accent hover:text-fg"
        >
          duration: (±1min)
        </button>
      </div>
      {isFetching && <div className="py-1 text-xs text-muted">Searching…</div>}
      <div className="max-h-[55vh] space-y-1.5 overflow-y-auto">
        {results.map((s) => (
          <div key={s.id} className="flex items-center gap-3 rounded-lg border border-line bg-surface-2 p-2">
            <img src={getImageURL(s.cover_url, '120x', sceneContext(s.scene_id))} alt="" className="h-14 w-14 shrink-0 rounded object-cover" loading="lazy" />
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-semibold">{s.title}</div>
              <div className="truncate text-xs text-muted">
                {s.site} · {formatDate(s.release_date)} · {s.duration}m · <span className="font-mono">{s.scene_id}</span>
              </div>
              <div className="truncate text-xs text-muted">{(s.cast ?? []).map((a) => a.name).join(', ')}</div>
              <div className="mt-1 h-1 overflow-hidden rounded-full bg-surface-3">
                <div className="h-full bg-accent" style={{ width: `${(((s._score ?? 0) / maxScore) * 100).toFixed(0)}%` }} />
              </div>
            </div>
            <button
              onClick={() => assign(s)}
              className="shrink-0 rounded-lg bg-accent px-3 py-1.5 text-xs font-semibold text-white"
            >
              Assign
            </button>
          </div>
        ))}
        {debounced && results.length === 0 && !isFetching && (
          <div className="py-6 text-center text-sm text-muted">No results</div>
        )}
      </div>
    </Modal>
  )
}

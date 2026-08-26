import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { Scene, SceneCuepoint } from '../../api/types'
import { humanizeSeconds1DP } from '../../lib/format'
import { StarRating } from '../../components/StarRating'
import { TrashIcon } from '../../components/icons'

// Cuepoints section: read-only table; editing controls in edit mode
// (track/name/start/end + quick-add chips + per-row rating/delete).
export function SceneCuepoints({
  scene,
  editMode,
  currentTime
}: {
  scene: Scene
  editMode: boolean
  currentTime: React.MutableRefObject<number>
}) {
  const queryClient = useQueryClient()
  const cuepoints = scene.cuepoints ?? []

  const { data: defaults } = useQuery({
    queryKey: ['cuepointDefaults'],
    queryFn: () => api.get<{ positions: string[]; actions: string[] }>('/options/cuepoints'),
    enabled: editMode
  })

  const [track, setTrack] = useState<string>('')
  const [name, setName] = useState('')
  const [start, setStart] = useState('')
  const [end, setEnd] = useState('')

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['scene'] })

  const addCue = useMutation({
    mutationFn: (body: Record<string, unknown>) => api.post(`/scene/${scene.id}/cuepoint`, body),
    onSuccess: () => {
      invalidate()
      setName('')
      setStart('')
      setEnd('')
    }
  })
  const delCue = useMutation({
    mutationFn: (cueId: number) => api.delete(`/scene/${scene.id}/cuepoint/${cueId}`),
    onSuccess: invalidate
  })
  // Per-row rating is stored by re-creating the cuepoint with a rating
  // (parity with the old UI).
  const rateCue = useMutation({
    mutationFn: async ({ cue, rating }: { cue: SceneCuepoint; rating: number }) => {
      await api.delete(`/scene/${scene.id}/cuepoint/${cue.id}`)
      await api.post(`/scene/${scene.id}/cuepoint`, {
        track: cue.track ?? null,
        name: cue.name,
        time_start: cue.time_start,
        time_end: cue.time_end ?? 0,
        rating
      })
    },
    onSuccess: invalidate
  })

  const parseTime = (v: string): number => {
    // accepts seconds or mm:ss or h:mm:ss(.s)
    if (/^\d+(\.\d+)?$/.test(v)) return parseFloat(v)
    const parts = v.split(':').map(Number)
    if (parts.some(isNaN)) return 0
    return parts.reduce((acc, p) => acc * 60 + p, 0)
  }

  const trackNum = track === '' ? null : Number(track)
  const endNum = end === '' ? 0 : parseTime(end)
  const valid = name.trim() !== '' && start !== '' && (trackNum === null || endNum > 0) && (end === '' || trackNum !== null)

  const add = () => {
    if (!valid) return
    addCue.mutate({
      track: trackNum,
      name: name.trim(),
      time_start: parseTime(start),
      time_end: endNum
    })
  }

  const quickAdd = (segment: string) => setName((n) => (n ? `${n}-${segment}` : segment))

  return (
    <div>
      {cuepoints.length === 0 && <div className="text-sm text-muted">No cuepoints.</div>}
      {cuepoints.length > 0 && (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-xs uppercase text-muted">
              <th className="py-1 pr-2">Track</th>
              <th className="py-1 pr-2">Name</th>
              <th className="py-1 pr-2">Start</th>
              <th className="py-1 pr-2">End</th>
              {editMode && <th className="py-1 pr-2">Rating</th>}
              {editMode && <th />}
            </tr>
          </thead>
          <tbody>
            {cuepoints.map((c) => (
              <tr key={c.id} className="border-t border-line">
                <td className="py-1 pr-2">{c.track ?? ''}</td>
                <td className="py-1 pr-2">{c.name}</td>
                <td className="py-1 pr-2 font-mono text-xs">{humanizeSeconds1DP(c.time_start)}</td>
                <td className="py-1 pr-2 font-mono text-xs">{c.time_end ? humanizeSeconds1DP(c.time_end) : ''}</td>
                {editMode && (
                  <td className="py-1 pr-2">
                    {c.track != null && (
                      <span className="flex items-center gap-1">
                        <StarRating
                          value={c.rating ?? 0}
                          size="sm"
                          onChange={(v) => rateCue.mutate({ cue: c, rating: v })}
                        />
                        {(c.rating ?? 0) > 0 && (
                          <button
                            onClick={() => rateCue.mutate({ cue: c, rating: 0 })}
                            className="text-xs text-muted hover:text-fg"
                            title="Reset rating"
                          >
                            ✕
                          </button>
                        )}
                      </span>
                    )}
                  </td>
                )}
                {editMode && (
                  <td className="py-1 text-right">
                    <button
                      onClick={() => delCue.mutate(c.id)}
                      className="text-muted hover:text-danger"
                      title="Delete cuepoint"
                    >
                      <TrashIcon />
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {editMode && (
        <div className="mt-3 rounded-lg border border-line bg-surface-2 p-3">
          <div className="flex flex-wrap items-end gap-2">
            <label className="text-xs">
              <span className="mb-0.5 block text-muted">Track</span>
              <input
                type="number"
                value={track}
                onChange={(e) => setTrack(e.target.value)}
                className="w-16 rounded border border-line bg-surface px-2 py-1"
              />
            </label>
            <label className="min-w-40 flex-1 text-xs">
              <span className="mb-0.5 block text-muted">Name</span>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                list="cuepoint-names"
                className="w-full rounded border border-line bg-surface px-2 py-1"
              />
              <datalist id="cuepoint-names">
                {[...(defaults?.positions ?? []), ...(defaults?.actions ?? [])].map((n) => (
                  <option key={n} value={n} />
                ))}
              </datalist>
            </label>
            <label className="text-xs">
              <span className="mb-0.5 block text-muted">Start</span>
              <input
                value={start}
                onChange={(e) => setStart(e.target.value)}
                placeholder="mm:ss"
                className="w-24 rounded border border-line bg-surface px-2 py-1 font-mono"
              />
            </label>
            <label className="text-xs">
              <span className="mb-0.5 block text-muted">End (HSP)</span>
              <input
                value={end}
                onChange={(e) => setEnd(e.target.value)}
                placeholder="mm:ss"
                className="w-24 rounded border border-line bg-surface px-2 py-1 font-mono"
              />
            </label>
            <button
              onClick={() => setStart(String(Math.floor(currentTime.current * 10) / 10))}
              className="rounded border border-line px-2 py-1 text-xs text-muted hover:text-fg"
              title="Use current player position"
            >
              Current time
            </button>
            <button
              onClick={add}
              disabled={!valid || addCue.isPending}
              className="rounded bg-accent px-3 py-1 text-xs font-semibold text-white disabled:opacity-50"
              title={trackNum !== null && endNum === 0 ? 'An end time is required for tracked (HSP) cuepoints' : ''}
            >
              Add
            </button>
          </div>
          <div className="mt-2 flex flex-wrap gap-1">
            {(defaults?.positions ?? []).map((p) => (
              <button key={p} onClick={() => quickAdd(p)} className="rounded-full bg-surface-3 px-2 py-0.5 text-[11px] hover:bg-accent-soft">
                {p}
              </button>
            ))}
            {(defaults?.actions ?? []).map((a) => (
              <button key={a} onClick={() => quickAdd(a)} className="rounded-full bg-accent-soft px-2 py-0.5 text-[11px] text-accent-strong">
                {a}
              </button>
            ))}
            {(scene.cast ?? []).map((a) => (
              <button key={a.id} onClick={() => quickAdd(a.name)} className="rounded-full bg-surface-3 px-2 py-0.5 text-[11px] italic hover:bg-accent-soft">
                {a.name}
              </button>
            ))}
            {(scene.tags ?? []).map((t) => (
              <button key={t.id} onClick={() => quickAdd(t.name)} className="rounded-full bg-surface-3 px-2 py-0.5 text-[11px] hover:bg-accent-soft">
                {t.name}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

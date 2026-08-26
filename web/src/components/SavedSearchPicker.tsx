import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Playlist } from '../api/types'
import { useToastStore } from '../store/toasts'
import { useUIStore } from '../store/ui'
import { Modal } from './Modal'
import { PencilIcon, PlusIcon, TrashIcon } from './icons'

// Saved-search picker / CRUD (server-side "playlists"). Shared by the scenes
// and actors filter popovers.
export function SavedSearchPicker({
  type,
  currentFilters,
  onApply
}: {
  type: 'scene' | 'actor'
  currentFilters: Record<string, unknown>
  onApply: (filters: Record<string, unknown>) => void
}) {
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const [selected, setSelected] = useState<number>(0)
  const [editOpen, setEditOpen] = useState(false)
  const [editName, setEditName] = useState('')
  const [editDeo, setEditDeo] = useState(false)
  const [isNew, setIsNew] = useState(false)

  const { data: playlists } = useQuery({
    queryKey: ['playlists', type],
    queryFn: () => api.get<Playlist[]>(type === 'actor' ? '/playlist/actor' : '/playlist')
  })

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['playlists'] })
    queryClient.invalidateQueries({ queryKey: ['sceneFilters'] })
    queryClient.invalidateQueries({ queryKey: ['actorFilters'] })
  }

  const save = useMutation({
    mutationFn: (body: Partial<Playlist> & { id?: number }) =>
      body.id ? api.put(`/playlist/${body.id}`, body) : api.post('/playlist', body),
    onSuccess: () => {
      useToastStore.getState().success('Saved search stored')
      invalidate()
      setEditOpen(false)
    }
  })

  const del = useMutation({
    mutationFn: (id: number) => api.delete(`/playlist/${id}`),
    onSuccess: () => {
      setSelected(0)
      invalidate()
    }
  })

  const list = playlists ?? []
  const current = list.find((p) => p.id === selected)

  const apply = (p: Playlist) => {
    setSelected(p.id)
    if (!p.search_params) return
    try {
      // Tolerant parse: old payloads carry client-only keys (cardSize etc.)
      onApply(JSON.parse(p.search_params))
    } catch {
      useToastStore.getState().error('Could not parse saved search')
    }
  }

  const openEditor = (isNew_: boolean) => {
    setIsNew(isNew_)
    setEditName(isNew_ ? '' : (current?.name ?? ''))
    setEditDeo(isNew_ ? false : (current?.is_deo_enabled ?? false))
    setEditOpen(true)
  }

  const submit = () => {
    const body: Record<string, unknown> = {
      name: editName,
      is_deo_enabled: editDeo,
      is_smart: true,
      search_params: JSON.stringify(currentFilters)
    }
    if (type === 'actor') body.playlist_type = 'actor'
    if (!isNew && current) {
      // Keep the stored params when just renaming? Old UI rewrites params on
      // edit — keep parity: edit saves current filters too.
      save.mutate({ ...body, id: current.id } as never)
    } else {
      save.mutate(body as never)
    }
  }

  const group = (deo: boolean) => list.filter((p) => p.is_deo_enabled === deo)

  return (
    <div>
      <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Saved searches</div>
      <div className="flex items-center gap-1">
        <select
          value={selected}
          onChange={(e) => {
            const p = list.find((x) => x.id === Number(e.target.value))
            if (p) apply(p)
            else setSelected(0)
          }}
          className="min-w-0 flex-1 rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
        >
          <option value={0}>—</option>
          {group(false).length > 0 && (
            <optgroup label="Web">
              {group(false).map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </optgroup>
          )}
          {group(true).length > 0 && (
            <optgroup label="VR Players">
              {group(true).map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </optgroup>
          )}
        </select>
        <button
          title="Save current filters as new saved search"
          onClick={() => openEditor(true)}
          className="rounded-lg border border-line px-2 py-1.5 text-sm hover:bg-surface-2"
        >
          <PlusIcon />
        </button>
        <button
          title="Edit selected"
          disabled={!current}
          onClick={() => openEditor(false)}
          className="rounded-lg border border-line px-2 py-1.5 text-sm hover:bg-surface-2 disabled:opacity-40"
        >
          <PencilIcon />
        </button>
        <button
          title="Delete selected"
          disabled={!current || current.is_system}
          onClick={async () => {
            if (current && (await askConfirm({ title: `Delete saved search "${current.name}"?`, danger: true }))) {
              del.mutate(current.id)
            }
          }}
          className="rounded-lg border border-line px-2 py-1.5 text-sm hover:bg-surface-2 disabled:opacity-40"
        >
          <TrashIcon />
        </button>
      </div>

      <Modal open={editOpen} onClose={() => setEditOpen(false)} width="max-w-sm" title={isNew ? 'Save search' : 'Edit saved search'}>
        <label className="mb-3 block">
          <span className="mb-1 block text-xs font-semibold uppercase text-muted">Name</span>
          <input
            autoFocus
            value={editName}
            onChange={(e) => setEditName(e.target.value)}
            className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
          />
        </label>
        {type === 'scene' && (
          <label className="mb-4 flex items-center gap-2 text-sm">
            <input type="checkbox" checked={editDeo} onChange={(e) => setEditDeo(e.target.checked)} />
            Use as DeoVR list
          </label>
        )}
        <div className="flex justify-end gap-2">
          <button onClick={() => setEditOpen(false)} className="rounded-lg border border-line px-3 py-1.5 text-sm">
            Cancel
          </button>
          <button
            disabled={!editName.trim() || save.isPending}
            onClick={submit}
            className="rounded-lg bg-accent px-3 py-1.5 text-sm text-white disabled:opacity-50"
          >
            Save
          </button>
        </div>
      </Modal>
    </div>
  )
}

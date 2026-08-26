import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { File, Scene } from '../api/types'
import { cleanFilenameForTitle } from '../lib/filename'
import { Modal } from './Modal'

// Create a custom scene from a file, then match the file to it.
export function CreateSceneModal({
  file,
  open,
  onClose,
  onDone
}: {
  file: File | null
  open: boolean
  onClose: () => void
  onDone: () => void
}) {
  const navigate = useNavigate()
  const [sceneId, setSceneId] = useState('')
  const [title, setTitle] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (open && file) {
      setSceneId('')
      setTitle(cleanFilenameForTitle(file))
    }
  }, [open, file])

  const create = async (andEdit: boolean) => {
    if (!file || !title.trim()) return
    setBusy(true)
    try {
      const scene = await api.post<Scene>('/scene/create', { title: title.trim(), id: sceneId.trim(), filename: file.filename })
      await api.post('/files/match', { file_id: file.id, scene_id: scene.scene_id })
      onDone()
      onClose()
      if (andEdit) navigate(`/scenes/${scene.scene_id}?edit=1`)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} width="max-w-md" title="Create custom scene">
      {file && <div className="mb-3 truncate font-mono text-xs text-muted">{file.filename}</div>}
      <label className="mb-3 block">
        <span className="mb-1 block text-xs font-semibold uppercase text-muted">Scene Id (optional)</span>
        <input
          value={sceneId}
          onChange={(e) => setSceneId(e.target.value)}
          placeholder="cannot be changed later"
          className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 font-mono text-sm"
        />
      </label>
      <label className="mb-4 block">
        <span className="mb-1 block text-xs font-semibold uppercase text-muted">Title</span>
        <input
          autoFocus
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
        />
      </label>
      <div className="flex justify-end gap-2">
        <button onClick={onClose} className="rounded-lg border border-line px-3 py-1.5 text-sm">
          Cancel
        </button>
        <button
          disabled={!title.trim() || busy}
          onClick={() => create(false)}
          className="rounded-lg border border-accent px-3 py-1.5 text-sm text-accent-strong disabled:opacity-50"
        >
          Create
        </button>
        <button
          disabled={!title.trim() || busy}
          onClick={() => create(true)}
          className="rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
        >
          Create and edit
        </button>
      </div>
    </Modal>
  )
}

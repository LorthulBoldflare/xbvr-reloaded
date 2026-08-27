import { useRef, useState } from 'react'
import type { SceneImage } from '../api/types'
import { getImageURL } from '../lib/image'
import { useUIStore } from '../store/ui'
import { Toggle } from './Toggle'
import { PlusIcon, TrashIcon } from './icons'

// Gallery editor: add by URL or by dropping local image files (data URLs),
// drag to reorder, delete (behind a lock), set-as-cover.
export function GalleryEditor({
  images,
  onChange,
  sceneId = '0'
}: {
  images: SceneImage[]
  onChange: (images: SceneImage[]) => void
  // scene_id of the scene being edited ('0' when unsaved/unknown)
  sceneId?: string
}) {
  const [newUrl, setNewUrl] = useState('')
  const [deleteUnlocked, setDeleteUnlocked] = useState(false)
  const dragIdx = useRef<number | null>(null)
  const askConfirm = useUIStore((s) => s.askConfirm)

  const add = (url: string, type: string = 'gallery') => {
    if (!url) return
    onChange([...images, { url, type }])
  }

  const addUrl = () => {
    add(newUrl.trim())
    setNewUrl('')
  }

  const onDropFiles = (e: React.DragEvent) => {
    e.preventDefault()
    for (const file of Array.from(e.dataTransfer.files)) {
      if (!file.type.startsWith('image/')) continue
      const reader = new FileReader()
      reader.onload = () => add(String(reader.result))
      reader.readAsDataURL(file)
    }
  }

  const remove = async (idx: number) => {
    const img = images[idx]
    if (img.type === 'cover') {
      const ok = await askConfirm({
        title: 'Delete the current cover image?',
        message: 'The scene will be left without a cover image.',
        danger: true
      })
      if (!ok) return
    }
    onChange(images.filter((_, i) => i !== idx))
  }

  const setCover = (idx: number) => {
    onChange(images.map((img, i) => ({ ...img, type: i === idx ? 'cover' : 'gallery' })))
  }

  const move = (from: number, to: number) => {
    if (from === to) return
    const next = [...images]
    const [item] = next.splice(from, 1)
    next.splice(to, 0, item)
    onChange(next)
  }

  return (
    <div>
      <div
        className="mb-2 flex gap-1"
        onDragOver={(e) => e.preventDefault()}
        onDrop={onDropFiles}
        title="Paste an image URL, or drop image files here"
      >
        <input
          value={newUrl}
          onChange={(e) => setNewUrl(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && addUrl()}
          placeholder="Add image URL (or drop files here)"
          className="min-w-0 flex-1 rounded-lg border border-dashed border-line-strong bg-surface-2 px-2 py-1.5 text-sm"
        />
        <button onClick={addUrl} className="rounded-lg border border-line px-2 py-1.5 text-sm hover:bg-surface-2">
          <PlusIcon />
        </button>
      </div>
      <Toggle checked={deleteUnlocked} onChange={setDeleteUnlocked} label="Unlock delete" />
      <div className="mt-2 grid grid-cols-[repeat(auto-fill,minmax(110px,1fr))] gap-2">
        {images.map((img, i) => (
          <div
            key={`${img.url.slice(0, 48)}-${i}`}
            draggable
            onDragStart={() => (dragIdx.current = i)}
            onDragOver={(e) => e.preventDefault()}
            onDrop={() => {
              if (dragIdx.current !== null) move(dragIdx.current, i)
              dragIdx.current = null
            }}
            className={`group relative cursor-move overflow-hidden rounded-lg border ${
              img.type === 'cover' ? 'border-accent ring-1 ring-accent' : 'border-line'
            }`}
          >
            <img src={getImageURL(img.url.replaceAll('\\', '/'), '200x', sceneId)} alt="" className="aspect-square w-full object-cover" />
            {img.type === 'cover' && (
              <span className="absolute left-1 top-1 rounded bg-accent px-1 text-[10px] font-bold text-white">COVER</span>
            )}
            <div className="absolute inset-x-0 bottom-0 flex justify-center gap-1 bg-black/60 p-1 opacity-0 transition-opacity group-hover:opacity-100">
              {img.type !== 'cover' && (
                <button onClick={() => setCover(i)} className="rounded bg-surface px-1.5 py-0.5 text-[10px] font-semibold">
                  Set as cover
                </button>
              )}
              {deleteUnlocked && (
                <button
                  onClick={() => remove(i)}
                  className="rounded bg-danger px-1.5 py-0.5 text-[10px] font-semibold text-white"
                >
                  <TrashIcon className="h-3 w-3" />
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

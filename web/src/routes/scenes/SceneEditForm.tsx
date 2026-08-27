import { useEffect, useState } from 'react'
import type { Scene, SceneImage } from '../../api/types'
import { GalleryEditor } from '../../components/GalleryEditor'
import { ListEditor } from '../../components/ListEditor'
import { TagInputEditor } from '../../components/TagInputEditor'
import { parseSceneImages } from './SceneGallery'
import { useSceneFilterOptions } from './FiltersPopover'

export interface SceneDraft {
  title: string
  studio: string
  site: string
  scene_url: string
  release_date_text: string
  duration: string
  is_multipart: boolean
  cast: string[]
  tags: string[]
  synopsis: string
  filenames: string[]
  images: SceneImage[]
}

export function draftFromScene(scene: Scene): SceneDraft {
  let filenames: string[] = []
  try {
    // note: the server may store the literal string "null" — JSON.parse
    // succeeds and returns null, so a catch alone is not sufficient
    const parsed = JSON.parse(scene.filenames_arr || '[]')
    if (Array.isArray(parsed)) filenames = parsed
  } catch {
    /* keep empty */
  }
  return {
    title: scene.title,
    studio: scene.studio,
    site: scene.site,
    scene_url: scene.scene_url,
    release_date_text: scene.release_date_text,
    duration: String(scene.duration ?? 0),
    is_multipart: scene.is_multipart,
    cast: (scene.cast ?? []).map((a) => a.name),
    tags: (scene.tags ?? []).map((t) => t.name),
    synopsis: scene.synopsis,
    filenames,
    images: parseSceneImages(scene.images)
  }
}

// Serialize the draft into the RequestEditSceneDetails body (cover first,
// deduped; cover_url mirrors the cover image — parity with the old UI).
export function draftToRequest(d: SceneDraft) {
  const seen = new Set<string>()
  const images = [...d.images]
    .sort((a, b) => (a.type === 'cover' ? -1 : b.type === 'cover' ? 1 : 0))
    .filter((img) => {
      if (seen.has(img.url)) return false
      seen.add(img.url)
      return true
    })
    .map((img) => ({ url: img.url, type: img.type, orientation: img.orientation ?? '' }))
  const cover = images.find((i) => i.type === 'cover')
  return {
    title: d.title,
    synopsis: d.synopsis,
    studio: d.studio,
    site: d.site,
    scene_url: d.scene_url,
    release_date_text: d.release_date_text,
    castArray: d.cast,
    tagsArray: d.tags,
    filenames_arr: JSON.stringify(d.filenames.filter((f) => f.trim() !== '')),
    images: JSON.stringify(images),
    cover_url: cover?.url ?? '',
    is_multipart: d.is_multipart,
    duration: d.duration
  }
}

export function SceneEditForm({
  draft,
  onChange,
  sceneId = '0'
}: {
  draft: SceneDraft
  onChange: (d: SceneDraft) => void
  // scene_id of the scene being edited ('0' when unsaved/unknown)
  sceneId?: string
}) {
  const { data: opts } = useSceneFilterOptions()
  const set = (p: Partial<SceneDraft>) => onChange({ ...draft, ...p })

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        <label className="md:col-span-2">
          <span className="mb-1 block text-xs font-semibold uppercase text-muted">Title</span>
          <input
            value={draft.title}
            onChange={(e) => set({ title: e.target.value })}
            className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
          />
        </label>
        <label>
          <span className="mb-1 block text-xs font-semibold uppercase text-muted">Studio</span>
          <input
            value={draft.studio}
            onChange={(e) => set({ studio: e.target.value })}
            className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
          />
        </label>
        <label>
          <span className="mb-1 block text-xs font-semibold uppercase text-muted">Site</span>
          <input
            value={draft.site}
            onChange={(e) => set({ site: e.target.value })}
            className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
          />
        </label>
        <label className="md:col-span-2">
          <span className="mb-1 block text-xs font-semibold uppercase text-muted">Scene URL</span>
          <input
            value={draft.scene_url}
            onChange={(e) => set({ scene_url: e.target.value })}
            className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 font-mono text-xs"
          />
        </label>
        <label>
          <span className="mb-1 block text-xs font-semibold uppercase text-muted">Release date</span>
          <input
            type="date"
            value={draft.release_date_text}
            onChange={(e) => set({ release_date_text: e.target.value })}
            className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
          />
        </label>
        <label>
          <span className="mb-1 block text-xs font-semibold uppercase text-muted">Duration (minutes)</span>
          <input
            type="number"
            value={draft.duration}
            onChange={(e) => set({ duration: e.target.value })}
            className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
          />
        </label>
        <label className="flex items-center gap-2 text-sm md:col-span-2">
          <input
            type="checkbox"
            checked={draft.is_multipart}
            onChange={(e) => set({ is_multipart: e.target.checked })}
          />
          Multipart scene
        </label>
      </div>

      <TagInputEditor label="Cast" values={draft.cast} options={opts?.cast ?? []} onChange={(v) => set({ cast: v })} />
      <TagInputEditor label="Tags" values={draft.tags} options={opts?.tags ?? []} onChange={(v) => set({ tags: v })} />

      <label>
        <span className="mb-1 block text-xs font-semibold uppercase text-muted">Description</span>
        <textarea
          value={draft.synopsis}
          onChange={(e) => set({ synopsis: e.target.value })}
          rows={5}
          className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
        />
      </label>

      <div>
        <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Filenames</div>
        <ListEditor items={draft.filenames} onChange={(v) => set({ filenames: v })} />
      </div>

      <div>
        <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Gallery</div>
        <GalleryEditor images={draft.images} onChange={(v) => set({ images: v })} sceneId={sceneId} />
      </div>
    </div>
  )
}

import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { Actor, Scene, SceneSearchField } from '../../api/types'
import { useOptionsState } from '../../api/hooks'
import { useUIStore } from '../../store/ui'
import { useToastStore } from '../../store/toasts'
import { useSceneFilterStore, applyDlState } from '../../store/sceneFilters'
import { patchSceneInCaches } from '../../api/sceneCache'
import { rescrapeScene } from '../../api/rescrape'
import { getImageURL } from '../../lib/image'
import { formatDate, formatDateTime, humanizeSeconds, safeHref } from '../../lib/format'
import { ScenePlayer } from '../../components/ScenePlayer'
import { StarRating } from '../../components/StarRating'
import { SceneFlagButtons, useSceneToggle } from '../../components/SceneFlagButtons'
import { SceneGallery, parseSceneImages } from './SceneGallery'
import { SceneFilesSection } from './SceneFilesSection'
import { SceneCuepoints } from './SceneCuepoints'
import { AltSourcesSection } from './AltSourcesSection'
import { SceneEditForm, draftFromScene, draftToRequest, type SceneDraft } from './SceneEditForm'
import type { File } from '../../api/types'

function ageInScene(actor: Actor, release: string): string {
  if (!actor.birth_date || actor.birth_date.startsWith('0001-01-01')) return ''
  if (!release || release.startsWith('0001-01-01')) return ''
  const age = Math.floor((Date.parse(release) - Date.parse(actor.birth_date)) / (365.25 * 24 * 3600 * 1000))
  return age > 0 && age < 100 ? ` (${age})` : ''
}

export function ScenePage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const queryClient = useQueryClient()
  const { data: state } = useOptionsState()
  const web = state?.config?.web
  const advanced = state?.config?.advanced
  const askConfirm = useUIStore((s) => s.askConfirm)
  const toast = useToastStore()

  const sceneQuery = useQuery({
    queryKey: ['scene', 'sid', id],
    queryFn: () => api.get<Scene>(`/scene/${encodeURIComponent(id)}`, { toastOnError: false }),
    retry: false
  })
  const scene = sceneQuery.data

  const [editMode, setEditMode] = useState(searchParams.get('edit') === '1')
  const [draft, setDraft] = useState<SceneDraft | null>(null)
  const [dirty, setDirty] = useState(false)
  const [playFile, setPlayFile] = useState<File | null>(null)
  const [showPreview, setShowPreview] = useState(false)
  const playerTime = useRef(0)
  const playerRef = useRef<HTMLDivElement>(null)

  const toggle = useSceneToggle()

  // Not found → toast + back to list.
  useEffect(() => {
    if (sceneQuery.isError || (scene && scene.id === 0)) {
      toast.error('Scene not found')
      navigate('/', { replace: true })
    }
  }, [sceneQuery.isError, scene, navigate, toast])

  const images = useMemo(() => (scene ? parseSceneImages(scene.images) : []), [scene])
  const videos = useMemo(() => (scene?.file ?? []).filter((f) => f.type === 'video'), [scene])
  const activeFile = playFile ?? videos[0] ?? null

  // Keep a fresh copy in the by-id cache for mutations.
  useEffect(() => {
    if (scene && scene.id !== 0) queryClient.setQueryData(['scene', scene.id], scene)
  }, [scene, queryClient])

  // Initialize the edit draft when edit mode is on (incl. direct ?edit=1 links).
  useEffect(() => {
    if (editMode && scene && scene.id !== 0 && !draft) {
      setDraft(draftFromScene(scene))
      setDirty(false)
    }
  }, [editMode, scene, draft])

  const rate = useMutation({
    mutationFn: (rating: number) => api.post<Scene>(`/scene/rate/${scene!.id}`, { rating }),
    onSuccess: (updated) => {
      queryClient.setQueryData(['scene', 'sid', id], updated)
      patchSceneInCaches(queryClient, updated.id, { star_rating: updated.star_rating })
    }
  })

  const save = useMutation({
    mutationFn: (d: SceneDraft) => api.post<Scene>(`/scene/edit/${scene!.id}`, draftToRequest(d)),
    onSuccess: (updated) => {
      queryClient.setQueryData(['scene', 'sid', id], updated)
      queryClient.invalidateQueries({ queryKey: ['sceneList'] })
      toast.success('Scene saved')
      setEditMode(false)
      setDirty(false)
    }
  })

  const deleteScene = useMutation({
    mutationFn: () => api.post('/scene/delete', { scene_id: scene!.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sceneList'] })
      navigate('/')
    }
  })

  const deletePreview = useMutation({
    mutationFn: () => api.delete<Scene>(`/scene/${scene!.id}/preview`),
    onSuccess: (updated) => queryClient.setQueryData(['scene', 'sid', id], updated)
  })

  const patchFiltersAndGoHome = (p: { cast?: string[]; sites?: string[]; tags?: string[] }) => {
    const store = useSceneFilterStore.getState()
    // Chip deep links land on the downloaded view ("Available right now"),
    // not the whole library.
    let f = applyDlState({ ...store.filters, cast: [], sites: [], tags: [] }, 'available')
    f = { ...f, ...p }
    store.setFilters(f)
    navigate('/')
  }

  // ---- keyboard shortcuts (page scope; suppressed while typing/modal open)
  const modalOpen = useUIStore((s) => s.confirm !== null || s.actorDetailsId !== null || s.quickFindOpen || s.stashdbSceneSearchId !== null)
  useEffect(() => {
    const order: string[] = JSON.parse(sessionStorage.getItem('sceneOrder') ?? '[]')
    const idx = order.indexOf(id)
    const go = (d: number) => {
      if (idx >= 0 && order[idx + d]) navigate(`/scenes/${order[idx + d]}`)
    }
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      if (['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) || target.isContentEditable) return
      if (modalOpen) return
      if (e.key === 'Escape') {
        if (editMode) {
          e.stopPropagation()
          tryExitEdit()
        } else navigate(-1)
      } else if (e.key === 'o') go(-1)
      else if (e.key === 'p') go(1)
      else if (e.key === 'e' && scene) enterEdit()
      else if (e.key === 'f' && scene) toggle.mutate({ scene, list: 'favourite' })
      else if (e.key === 'w' && scene) toggle.mutate({ scene, list: 'watchlist' })
      else if (e.key === 'W' && scene) toggle.mutate({ scene, list: 'watched' })
      else if (e.key === 't' && scene) toggle.mutate({ scene, list: 'trailerlist' })
      else if (e.key === '0' && scene) rate.mutate(0)
      else if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
        window.dispatchEvent(new CustomEvent('scene-gallery-nav', { detail: e.key === 'ArrowLeft' ? -1 : 1 }))
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, editMode, dirty, modalOpen, scene, draft])

  const enterEdit = () => {
    if (!scene) return
    setDraft(draftFromScene(scene))
    setDirty(false)
    setEditMode(true)
  }

  const tryExitEdit = async () => {
    if (dirty) {
      const ok = await askConfirm({ title: 'Discard unsaved changes?' })
      if (!ok) return
    }
    setEditMode(false)
    setDirty(false)
    if (searchParams.get('edit')) {
      setSearchParams((prev) => {
        const n = new URLSearchParams(prev)
        n.delete('edit')
        return n
      }, { replace: true })
    }
  }

  if (!scene || scene.id === 0) {
    return <div className="py-16 text-center text-muted">Loading…</div>
  }

  const setDraftDirty = (d: SceneDraft) => {
    setDraft(d)
    setDirty(true)
  }

  return (
    <div>
      {/* hero: blurred cover backdrop, poster, title, meta, rating, actions */}
      <div className="relative mb-6 overflow-hidden rounded-2xl border border-line">
        <img
          src={getImageURL(scene.cover_url)}
          alt=""
          aria-hidden
          className="absolute inset-0 h-full w-full scale-125 object-cover opacity-25 blur-2xl"
          onError={(e) => ((e.target as HTMLImageElement).style.display = 'none')}
        />
        <div className="absolute inset-0 bg-gradient-to-r from-page/90 via-page/60 to-transparent" />
        <div className="relative flex flex-wrap items-center gap-5 p-5 lg:p-7">
          <img
            src={getImageURL(scene.cover_url)}
            alt=""
            onError={(e) => {
              const el = e.target as HTMLImageElement
              if (!el.src.endsWith('blank.png')) el.src = `${import.meta.env.BASE_URL}blank.png`
            }}
            className="hidden w-52 shrink-0 rounded-xl object-cover shadow-2xl ring-1 ring-line-strong sm:block"
          />
          <div className="min-w-0 flex-1">
            <h1 className="text-2xl font-extrabold leading-tight lg:text-3xl">{scene.title}</h1>
            <div className="mt-2 flex flex-wrap items-center gap-1.5 text-xs">
              {scene.scene_url && (
                <a
                  href={safeHref(scene.scene_url)}
                  target="_blank"
                  rel="noreferrer"
                  className={`rounded-full px-2.5 py-1 font-semibold ${scene.is_subscribed ? 'bg-accent text-accent-fg' : 'bg-surface-3 hover:bg-accent-soft'}`}
                >
                  {scene.site}
                </a>
              )}
              {scene.members_url && (
                <a
                  href={safeHref(scene.members_url)}
                  target="_blank"
                  rel="noreferrer"
                  className="rounded-full bg-surface-3 px-2.5 py-1 hover:bg-accent-soft"
                >
                  Members
                </a>
              )}
              <span className="rounded-full bg-surface-3 px-2.5 py-1">{formatDate(scene.release_date)}</span>
              {scene.duration > 0 && <span className="rounded-full bg-surface-3 px-2.5 py-1">{scene.duration} min</span>}
              {scene.is_hidden && <span className="rounded-full bg-warn/20 px-2.5 py-1 text-warn">hidden</span>}
            </div>
            <div className="mt-3 flex flex-wrap items-center gap-3">
              <span className="flex items-center gap-1">
                <StarRating value={scene.star_rating} onChange={(v) => rate.mutate(v)} size="lg" />
                {scene.star_rating > 0 && (
                  <button onClick={() => rate.mutate(0)} className="text-xs text-muted hover:text-fg" title="Reset rating (0)">
                    ✕
                  </button>
                )}
              </span>
              <SceneFlagButtons scene={scene} onEdit={enterEdit} />
            </div>
          </div>
          <div className="shrink-0">
            {!editMode ? (
              <button
                onClick={enterEdit}
                className="btn-gradient rounded-full px-5 py-2 text-sm font-bold shadow-lg transition-transform hover:scale-105"
              >
                Edit
              </button>
            ) : (
              <div className="flex gap-2">
                <button onClick={tryExitEdit} className="rounded-full border border-line bg-surface px-4 py-2 text-sm font-semibold">
                  Cancel
                </button>
                <button
                  onClick={() => draft && save.mutate(draft)}
                  disabled={save.isPending}
                  className="rounded-full bg-ok px-5 py-2 text-sm font-bold text-white shadow-lg disabled:opacity-50"
                >
                  Save
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* main two-column area */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div>
          {videos.length > 0 && (
            <div ref={playerRef} className="mb-4">
              <ScenePlayer file={activeFile} poster={scene.cover_url} onTimeUpdate={(t) => (playerTime.current = t)} />
              {scene.has_preview && (
                <div className="mt-1.5 text-center">
                  <button
                    onClick={() => setShowPreview((s) => !s)}
                    className="rounded-lg border border-line px-3 py-1 text-xs text-muted hover:text-fg"
                  >
                    {showPreview ? 'Hide preview' : 'Play preview'}
                  </button>
                  {showPreview && (
                    <video
                      src={`/api/dms/preview/${scene.scene_id}`}
                      controls
                      autoPlay
                      loop
                      muted
                      playsInline
                      className="mt-2 w-full rounded-xl bg-black"
                    />
                  )}
                </div>
              )}
            </div>
          )}
          {videos.length === 0 && scene.has_preview && (
            <video
              src={`/api/dms/preview/${scene.scene_id}`}
              controls
              loop
              muted
              playsInline
              poster={getImageURL(scene.cover_url, '700,fit')}
              className="mb-4 w-full rounded-xl bg-black"
            />
          )}
          <SceneGallery images={images} coverUrl={scene.cover_url} />
        </div>

        <div className="space-y-5">
          {/* cast — round thumbnails */}
          {(scene.cast ?? []).length > 0 && (
            <section>
              <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted">Cast</h2>
              <div className="flex flex-wrap gap-3">
                {(scene.cast ?? []).map((a) => (
                  <button
                    key={a.id}
                    onClick={() => useUIStore.getState().showActorDetails(a.id)}
                    className="group w-20 text-center"
                    title={a.name}
                  >
                    <img
                      src={getImageURL(a.image_url, '200x')}
                      alt={a.name}
                      loading="lazy"
                      onError={(e) => {
                        ;(e.target as HTMLImageElement).src = `${import.meta.env.BASE_URL}blank_female_profile.png`
                      }}
                      className="mx-auto h-16 w-16 rounded-full border-2 border-line object-cover transition-colors group-hover:border-accent"
                    />
                    <span className="mt-1 block truncate text-xs">
                      {a.name}
                      <span className="text-muted">{ageInScene(a, scene.release_date)}</span>
                    </span>
                  </button>
                ))}
              </div>
            </section>
          )}

          {/* tags */}
          <section className="space-y-1.5">
            <div className="flex flex-wrap gap-1">
              {(scene.cast ?? []).map((a) => (
                <button
                  key={`c${a.id}`}
                  onClick={() => patchFiltersAndGoHome({ cast: [a.name] })}
                  className="rounded-full bg-accent-soft px-2 py-0.5 text-xs text-accent-strong hover:brightness-110"
                  title={`Filter scenes by ${a.name}`}
                >
                  {a.name}
                  {a.count ? <span className="ml-1 text-muted">{a.avail_count}/{a.count}</span> : null}
                </button>
              ))}
              <button
                onClick={() => patchFiltersAndGoHome({ sites: [scene.site] })}
                className="rounded-full bg-surface-3 px-2 py-0.5 text-xs hover:bg-accent-soft"
              >
                {scene.site}
              </button>
              {[...(scene.tags ?? [])]
                .sort((a, b) => (web?.tagSort === 'alphabetically' ? a.name.localeCompare(b.name) : (b.count ?? 0) - (a.count ?? 0)))
                .map((t) => (
                  <button
                    key={t.id}
                    onClick={() => patchFiltersAndGoHome({ tags: [t.name] })}
                    className="rounded-full bg-surface-3 px-2 py-0.5 text-xs hover:bg-accent-soft"
                  >
                    {t.name}
                    {t.count ? <span className="ml-1 text-muted">{t.count}</span> : null}
                  </button>
                ))}
            </div>
          </section>

          {scene.synopsis && (
            <section>
              <h2 className="mb-1 text-sm font-semibold uppercase tracking-wide text-muted">Description</h2>
              <p className="whitespace-pre-line rounded-lg bg-surface-2 p-3 text-sm">{scene.synopsis}</p>
            </section>
          )}

          <AltSourcesSection scene={scene} />
        </div>
      </div>

      {/* edit form */}
      {editMode && draft && (
        <section className="mt-6 rounded-xl border border-accent/40 bg-surface p-4">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-muted">Edit scene</h2>
          <SceneEditForm draft={draft} onChange={setDraftDirty} />

          <div className="mt-4 flex flex-wrap gap-2 border-t border-line pt-3">
            <button
              onClick={() => toggle.mutate({ scene, list: 'needs_update' })}
              className={`rounded-lg border px-3 py-1.5 text-xs ${scene.needs_update ? 'border-warn text-warn' : 'border-line text-muted'}`}
            >
              {scene.needs_update ? 'Marked for update' : 'Mark for update'}
            </button>
            <button
              onClick={() => rescrapeScene(scene, () => toggle.mutate({ scene, list: 'needs_update' }))}
              className="rounded-lg border border-line px-3 py-1.5 text-xs text-muted hover:text-fg"
            >
              Rescrape scene
            </button>
            {scene.has_preview && (
              <button
                onClick={async () => {
                  if (await askConfirm({ title: 'Delete the generated preview video?', danger: true })) deletePreview.mutate()
                }}
                className="rounded-lg border border-line px-3 py-1.5 text-xs text-muted hover:text-fg"
              >
                Delete preview
              </button>
            )}
            <button
              onClick={() => useUIStore.getState().showStashdbSceneSearch(scene.id)}
              className="rounded-lg border border-line px-3 py-1.5 text-xs text-muted hover:text-fg"
            >
              Link StashDB…
            </button>
            <button
              onClick={async () => {
                if (
                  await askConfirm({
                    title: 'Delete this scene?',
                    message: 'The scene will be re-added on the next scrape unless the source is removed.',
                    danger: true,
                    confirmLabel: 'Delete'
                  })
                )
                  deleteScene.mutate()
              }}
              className="ml-auto rounded-lg border border-danger/40 px-3 py-1.5 text-xs text-danger hover:bg-danger/10"
            >
              Delete scene
            </button>
          </div>
        </section>
      )}

      {/* files / cuepoints / history */}
      <div className="mt-6 space-y-6">
        <section>
          <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted">
            Files ({(scene.file ?? []).length})
          </h2>
          <SceneFilesSection
            scene={scene}
            editMode={editMode}
            onPlay={(f) => {
              setPlayFile(f)
              playerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
            }}
          />
        </section>

        <section>
          <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted">
            Cuepoints ({(scene.cuepoints ?? []).length})
          </h2>
          <SceneCuepoints scene={scene} editMode={editMode} currentTime={playerTime} />
        </section>

        {(scene.history ?? []).length > 0 && (
          <section>
            <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted">Watch history</h2>
            <div className="text-sm text-muted">
              {scene.history!.length} sessions · {humanizeSeconds(scene.history!.reduce((a, h) => a + h.duration, 0))} total
            </div>
            <ul className="mt-1 space-y-0.5 text-xs text-muted">
              {scene.history!.map((h) => (
                <li key={h.id}>
                  {formatDateTime(h.time_start)} — {humanizeSeconds(h.duration)}
                </li>
              ))}
            </ul>
          </section>
        )}

        {advanced?.showSceneSearchField && <SearchFields sceneId={scene.id} />}
      </div>

      {/* footer */}
      <footer className="mt-8 flex flex-wrap items-center gap-3 border-t border-line pt-3 text-xs text-muted">
        <span className="font-mono">{scene.scene_id}</span>
        {advanced?.showInternalSceneId && <span>internal id: {scene.id}</span>}
        {advanced?.showHSPApiLink && (
          <a href={`/heresphere/${scene.id}`} target="_blank" rel="noreferrer" className="hover:text-accent">
            HereSphere API
          </a>
        )}
      </footer>
    </div>
  )
}

function SearchFields({ sceneId }: { sceneId: number }) {
  const { data } = useQuery({
    queryKey: ['sceneSearchFields', sceneId],
    queryFn: () => api.get<SceneSearchField[]>(`/scene/searchfields?q=${sceneId}`)
  })
  if (!data || data.length === 0) return null
  return (
    <section>
      <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted">Search fields</h2>
      <table className="w-full text-xs">
        <tbody>
          {data.map((f) => (
            <tr key={f.fieldName} className="border-t border-line">
              <td className="py-1 pr-2 font-mono text-muted">{f.fieldName}</td>
              <td className="py-1">{f.fieldValue}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}

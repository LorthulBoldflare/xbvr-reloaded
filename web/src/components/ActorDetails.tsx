import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Actor, AkaResponse, CountryDetails, ExternalReferenceLink, Scene } from '../api/types'
import { getImageURL } from '../lib/image'
import { formatDate, safeHref } from '../lib/format'
import { useUIStore } from '../store/ui'
import { useToastStore } from '../store/toasts'
import { useOptionsState } from '../api/hooks'
import { Modal } from './Modal'
import { StarRating } from './StarRating'
import { SceneCard } from './SceneCard'
import { ActorCard } from './ActorCard'
import { LinkIcon, PencilIcon, TrashIcon } from './icons'

function parseJsonArray<T>(json: string): T[] {
  try {
    const arr = JSON.parse(json || '[]')
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

function age(birthDate: string): number | null {
  if (!birthDate || birthDate.startsWith('0001-01-01')) return null
  const years = Math.floor((Date.now() - Date.parse(birthDate)) / (365.25 * 24 * 3600 * 1000))
  return years > 0 && years < 120 ? years : null
}

// Global actor details modal (old UI kept this as an overlay — kept as a
// modal; only scene details became a page).
export function ActorDetails() {
  const actorId = useUIStore((s) => s.actorDetailsId)
  const hide = useUIStore((s) => s.hideActorDetails)
  const showEdit = useUIStore((s) => s.showEditActor)
  const showStashdb = useUIStore((s) => s.showStashdbActorSearch)
  const askConfirm = useUIStore((s) => s.askConfirm)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const toast = useToastStore()
  const { data: state } = useOptionsState()
  const web = state?.config?.web

  const [tab, setTab] = useState(0)
  const [imgIdx, setImgIdx] = useState(0)

  const { data: actor } = useQuery({
    queryKey: ['actor', actorId],
    queryFn: () => api.get<Actor>(`/actor/${actorId}`),
    enabled: actorId !== null
  })

  const { data: countries } = useQuery({
    queryKey: ['countrylist'],
    queryFn: () => api.get<CountryDetails[]>('/actor/countrylist'),
    staleTime: Infinity,
    enabled: actorId !== null
  })

  const { data: akas } = useQuery({
    queryKey: ['actorAkas', actorId],
    queryFn: () => api.get<AkaResponse>(`/actor/akas/${actorId}`),
    enabled: actorId !== null && tab === 2
  })
  const { data: colleagues } = useQuery({
    queryKey: ['actorColleagues', actorId],
    queryFn: () => api.get<Actor[]>(`/actor/colleagues/${actorId}`),
    enabled: actorId !== null && tab === 3
  })
  const { data: extrefs } = useQuery({
    queryKey: ['actorExtrefs', actorId],
    queryFn: () => api.get<ExternalReferenceLink[]>(`/actor/extrefs/${actorId}`),
    enabled: actorId !== null && tab === 5
  })

  useEffect(() => {
    setTab(0)
    setImgIdx(0)
  }, [actorId])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['actor', actorId] })
    queryClient.invalidateQueries({ queryKey: ['actorList'] })
    queryClient.invalidateQueries({ queryKey: ['actorAkas', actorId] })
    queryClient.invalidateQueries({ queryKey: ['actorExtrefs', actorId] })
  }

  const rate = useMutation({
    mutationFn: (rating: number) => api.post(`/actor/rate/${actorId}`, { rating }),
    onSuccess: invalidate
  })
  const toggle = useMutation({
    mutationFn: (list: 'watchlist' | 'favourite') => api.post('/actor/toggle', { actor_id: actorId, list }),
    onSuccess: invalidate
  })
  const setImage = useMutation({
    mutationFn: (url: string) => api.post('/actor/setimage', { actor_id: actorId, url }),
    onSuccess: invalidate
  })
  const delImage = useMutation({
    mutationFn: (url: string) => api.delete('/actor/delimage', { actor_id: actorId, url }),
    onSuccess: invalidate
  })

  // keyboard: esc close, ←/→ gallery, o/p prev/next actor, f/w toggles, e edit, s stashdb, 0 reset
  useEffect(() => {
    if (actorId === null) return
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      if (['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) || target.isContentEditable) return
      const order: number[] = JSON.parse(sessionStorage.getItem('actorOrder') ?? '[]')
      const idx = order.indexOf(actorId)
      if (e.key === 'ArrowLeft') setImgIdx((i) => Math.max(0, i - 1))
      else if (e.key === 'ArrowRight') setImgIdx((i) => i + 1)
      else if (e.key === 'o' && idx > 0) useUIStore.getState().showActorDetails(order[idx - 1])
      else if (e.key === 'p' && idx >= 0 && idx < order.length - 1) useUIStore.getState().showActorDetails(order[idx + 1])
      else if (e.key === 'f') toggle.mutate('favourite')
      else if (e.key === 'w') toggle.mutate('watchlist')
      else if (e.key === 'e') showEdit(actorId)
      else if (e.key === 's') showStashdb(actorId)
      else if (e.key === '0') rate.mutate(0)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [actorId])

  const images = useMemo(() => parseJsonArray<string>(actor?.image_arr ?? ''), [actor])
  const urls = useMemo(() => parseJsonArray<{ url: string; type: string }>(actor?.urls ?? ''), [actor])
  const aliases = useMemo(() => parseJsonArray<string>(actor?.aliases ?? ''), [actor])
  const scenes: Scene[] = actor?.scenes ?? []

  if (actorId === null) return null
  if (!actor) return <Modal open onClose={hide} width="max-w-4xl"><div className="py-10 text-center text-muted">Loading…</div></Modal>

  const isAka = actor.name.startsWith('aka:')
  const actorAge = age(actor.birth_date)
  const country = countries?.find((c) => c.name === actor.nationality)
  const tabs = [
    'Details',
    `Scenes (${scenes.length})`,
    'Akas',
    `Colleagues`,
    `Links (${urls.length})`,
    `Scrapers`
  ]

  const akaAction = async (op: 'create' | 'add' | 'remove' | 'delete', names: string[]) => {
    const { akaApi } = await import('../api/groups')
    let data
    if (op === 'create') data = await akaApi.create(names)
    else if (op === 'add') data = await akaApi.add(names)
    else if (op === 'remove') data = await akaApi.remove(names)
    else data = await akaApi.delete(names[0])
    if (data.status) toast.info(`Warning: ${data.status}`)
    invalidate()
    queryClient.invalidateQueries({ queryKey: ['sceneFilters'] })
  }

  return (
    <Modal open onClose={hide} width="max-w-5xl">
      <div className="grid grid-cols-1 gap-5 md:grid-cols-[240px_1fr]">
        {/* gallery */}
        <div>
          <div className="relative overflow-hidden rounded-xl bg-surface-3">
            <img
              src={getImageURL(images[imgIdx] ?? actor.image_url, '700x', 'act-' + actor.id)}
              alt={actor.name}
              onError={(e) => {
                const el = e.target as HTMLImageElement
                if (!el.src.endsWith('blank_female_profile.png'))
                  el.src = `${import.meta.env.BASE_URL}blank_female_profile.png`
              }}
              className="aspect-[2/3] w-full object-cover"
            />
          </div>
          {images.length > 1 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {images.map((u, i) => (
                <button key={i} onClick={() => setImgIdx(i)}>
                  <img
                    src={getImageURL(u, 'x85', 'act-' + actor.id)}
                    alt=""
                    className={`h-10 rounded object-cover ${i === imgIdx ? 'ring-2 ring-accent' : 'opacity-60 hover:opacity-100'}`}
                  />
                </button>
              ))}
            </div>
          )}
          {images.length > 0 && (
            <div className="mt-2 flex gap-1.5">
              <button
                onClick={() => setImage.mutate(images[imgIdx])}
                className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2"
              >
                Set main image
              </button>
              <button
                onClick={async () => {
                  if (await askConfirm({ title: 'Delete this image?', danger: true })) delImage.mutate(images[imgIdx])
                }}
                className="rounded-lg border border-line px-2 py-1 text-xs text-muted hover:text-danger"
              >
                <TrashIcon />
              </button>
            </div>
          )}
        </div>

        {/* content */}
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-lg font-bold">{actor.name}</h2>
            {isAka && (
              <button
                onClick={async () => {
                  if (await askConfirm({ title: `Delete aka group ${actor.name}?`, danger: true })) {
                    await akaAction('delete', [actor.name])
                    hide()
                  }
                }}
                className="rounded-lg border border-danger/40 px-2 py-0.5 text-xs text-danger"
              >
                Delete group
              </button>
            )}
            {!isAka && (akas?.possible_akas?.length ?? 0) > 0 && (
              <button
                onClick={async () => {
                  if (
                    await askConfirm({
                      title: 'Create aka group?',
                      message: `Group ${actor.name} with: ${akas!.possible_akas.map((p) => p.name).join(', ')}?`
                    })
                  )
                    akaAction('create', [actor.name, ...akas!.possible_akas.map((p) => p.name)])
                }}
                className="rounded-lg border border-line px-2 py-0.5 text-xs text-muted hover:text-fg"
              >
                Create aka group
              </button>
            )}
          </div>
          <div className="mt-1 flex flex-wrap gap-1.5 text-xs text-muted">
            {actor.birth_date && !actor.birth_date.startsWith('0001-01-01') && (
              <span className="rounded-full bg-surface-3 px-2 py-0.5">
                {formatDate(actor.birth_date)}
                {actorAge !== null && ` (${actorAge})`}
              </span>
            )}
            {(actor.start_year > 0 || actor.end_year > 0) && (
              <span className="rounded-full bg-surface-3 px-2 py-0.5">
                active {actor.start_year || '?'}–{actor.end_year || ''}
              </span>
            )}
            <span className="rounded-full bg-surface-3 px-2 py-0.5">
              {actor.avail_count}/{actor.count} scenes
            </span>
          </div>

          <div className="mt-2 flex items-center gap-3">
            <span className="flex items-center gap-1">
              <StarRating value={actor.star_rating} onChange={(v) => rate.mutate(v)} />
              {actor.star_rating > 0 && (
                <button onClick={() => rate.mutate(0)} className="text-xs text-muted hover:text-fg" title="Reset rating">
                  ✕
                </button>
              )}
            </span>
            {Number(actor.scene_rating_average) > 0 && (
              <span className="text-xs text-muted">scene avg: {Number(actor.scene_rating_average).toFixed(2)}</span>
            )}
          </div>

          <div className="mt-2 flex flex-wrap gap-1.5">
            <button
              onClick={() => toggle.mutate('favourite')}
              className={`rounded-lg border px-2.5 py-1 text-xs ${actor.favourite ? 'border-accent text-accent-strong' : 'border-line text-muted'}`}
            >
              ♥ favourite
            </button>
            <button
              onClick={() => toggle.mutate('watchlist')}
              className={`rounded-lg border px-2.5 py-1 text-xs ${actor.watchlist ? 'border-accent text-accent-strong' : 'border-line text-muted'}`}
            >
              watchlist
            </button>
            <button
              onClick={() => showEdit(actor.id)}
              className="rounded-lg border border-line px-2.5 py-1 text-xs text-muted hover:text-fg"
            >
              <span className="flex items-center gap-1"><PencilIcon /> edit</span>
            </button>
            <button
              onClick={() => showStashdb(actor.id)}
              className="rounded-lg border border-line px-2.5 py-1 text-xs text-muted hover:text-fg"
            >
              Link StashDB…
            </button>
          </div>

          {/* tabs */}
          <div className="mt-3 flex gap-1 border-b border-line">
            {tabs.map((t, i) => (
              <button
                key={t}
                onClick={() => setTab(i)}
                className={`-mb-px border-b-2 px-2.5 py-1.5 text-xs font-medium ${
                  tab === i ? 'border-accent text-accent-strong' : 'border-transparent text-muted hover:text-fg'
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          <div className="pt-3">
            {tab === 0 && (
              <table className="w-full text-sm">
                <tbody>
                  {[
                    ['Age', actorAge !== null ? String(actorAge) : ''],
                    ['Nationality', country ? `${country.name}` : actor.nationality],
                    ['Ethnicity', actor.ethnicity],
                    ['Hair', actor.hair_color],
                    ['Eyes', actor.eye_color],
                    actor.height > 0 && ['Height', `${actor.height} cm (${Math.floor(actor.height / 2.54 / 12)}'${Math.round((actor.height / 2.54) % 12)}")`],
                    actor.weight > 0 && ['Weight', `${actor.weight} kg (${Math.round(actor.weight * 2.20462)} lbs)`],
                    (actor.band_size > 0 || actor.cup_size) && [
                      'Measurements',
                      `${actor.band_size}${actor.cup_size}-${actor.waist_size}-${actor.hip_size}`
                    ],
                    ['Breast type', actor.breast_type],
                    aliases.length > 0 && ['Aliases', aliases.join(', ')],
                    ['Tattoos', actor.tattoos],
                    ['Piercings', actor.piercings]
                  ]
                    .filter(Boolean)
                    .map((row) => {
                      const [k, v] = row as [string, string]
                      return (
                        <tr key={k} className="border-t border-line">
                          <td className="w-32 py-1.5 pr-2 text-xs font-semibold uppercase text-muted">{k}</td>
                          <td className="py-1.5">{v}</td>
                        </tr>
                      )
                    })}
                </tbody>
              </table>
            )}
            {tab === 0 && actor.biography && (
              <p className="mt-2 whitespace-pre-line rounded-lg bg-surface-2 p-3 text-sm">{actor.biography}</p>
            )}

            {tab === 1 && (
              <div className="grid max-h-[50vh] grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-3 overflow-y-auto">
                {scenes.map((s) => (
                  <SceneCard
                    key={s.id}
                    scene={s}
                    onOpen={(scene) => {
                      hide()
                      navigate(`/scenes/${scene.scene_id}`)
                    }}
                  />
                ))}
                {scenes.length === 0 && <div className="text-sm text-muted">No scenes</div>}
              </div>
            )}

            {tab === 2 && (
              <div className="space-y-3">
                {(akas?.aka_groups ?? []).length > 0 && (
                  <div>
                    <div className="mb-1 text-xs font-semibold uppercase text-muted">Groups</div>
                    <div className="flex flex-wrap gap-1">
                      {akas!.aka_groups.map((g) => (
                        <span key={g.id} className="rounded-full bg-accent-soft px-2 py-0.5 text-xs text-accent-strong">
                          {g.name}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                {(akas?.actors ?? []).length > 0 && (
                  <div>
                    <div className="mb-1 text-xs font-semibold uppercase text-muted">Members</div>
                    <div className="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-3">
                      {akas!.actors.map((a) => (
                        <div key={a.id} className="relative">
                          <ActorCard actor={a} />
                          <button
                            onClick={async () => {
                              if (await askConfirm({ title: `Remove ${a.name} from the group?` }))
                                akaAction('remove', [actor.name, a.name])
                            }}
                            className="absolute right-1 top-1 rounded-full bg-black/60 p-1 text-white hover:bg-danger"
                            title="Remove from group"
                          >
                            <TrashIcon className="h-3 w-3" />
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
                {(akas?.possible_akas ?? []).length > 0 && (
                  <div>
                    <div className="mb-1 text-xs font-semibold uppercase text-muted">Possible matches</div>
                    <div className="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-3">
                      {akas!.possible_akas.map((a) => (
                        <div key={a.id} className="relative">
                          <ActorCard actor={a} />
                          <button
                            onClick={() => akaAction('add', [actor.name, a.name])}
                            className="absolute right-1 top-1 rounded-full bg-black/60 px-1.5 py-0.5 text-[10px] font-bold text-white hover:bg-ok"
                            title="Add to group"
                          >
                            + add
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {tab === 3 && (
              <div className="grid max-h-[50vh] grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-3 overflow-y-auto">
                {(colleagues ?? []).map((a) => (
                  <ActorCard key={a.id} actor={a} colleagueMode onClick={() => useUIStore.getState().showActorDetails(a.id)} />
                ))}
                {(colleagues ?? []).length === 0 && <div className="text-sm text-muted">No colleagues</div>}
              </div>
            )}

            {tab === 4 && (
              <div className="space-y-1">
                {urls.map((u, i) => (
                  <a
                    key={i}
                    href={safeHref(u.url)}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center gap-2 rounded-lg border border-line px-3 py-1.5 text-sm hover:bg-surface-2"
                  >
                    <LinkIcon /> {u.url}
                    {u.type && <span className="text-xs text-muted">({u.type})</span>}
                  </a>
                ))}
                {urls.length === 0 && <div className="text-sm text-muted">No links</div>}
              </div>
            )}

            {tab === 5 && (
              <div className="space-y-1">
                {(extrefs ?? []).map((r) => (
                  <div key={r.id} className="flex items-center gap-2 rounded-lg border border-line px-3 py-1.5 text-sm">
                    <a href={safeHref(r.url)} target="_blank" rel="noreferrer" className="min-w-0 flex-1 truncate hover:text-accent">
                      {r.external_source}
                    </a>
                    <button
                      title="Refresh from source"
                      onClick={async () => {
                        if (r.url.includes('stashdb.org')) {
                          await api.get(`/extref/stashdb/refresh_performer/${r.external_id}`)
                        } else {
                          await api.post('/extref/generic/scrape_single', { id: actor.id, url: r.url })
                        }
                        toast.success('Refresh requested')
                        invalidate()
                      }}
                      className="rounded border border-line px-2 py-0.5 text-xs text-muted hover:text-fg"
                    >
                      refresh
                    </button>
                  </div>
                ))}
                {(extrefs ?? []).length === 0 && <div className="text-sm text-muted">No scraper links</div>}
              </div>
            )}
          </div>
        </div>
      </div>
    </Modal>
  )
}

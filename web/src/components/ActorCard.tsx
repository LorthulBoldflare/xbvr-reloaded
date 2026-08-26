import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Actor } from '../api/types'
import { useOptionsState } from '../api/hooks'
import { getImageURL } from '../lib/image'
import { formatDate } from '../lib/format'
import { useUIStore } from '../store/ui'
import { StarRating } from './StarRating'

// Actor portrait card (actors grid + actor modal tabs).
export function ActorCard({
  actor,
  colleagueMode = false,
  onClick
}: {
  actor: Actor
  colleagueMode?: boolean
  onClick?: () => void
}) {
  const { data: state } = useOptionsState()
  const web = state?.currentState?.web
  const queryClient = useQueryClient()
  const showActorDetails = useUIStore((s) => s.showActorDetails)

  const toggle = useMutation({
    mutationFn: (list: 'watchlist' | 'favourite') => api.post('/actor/toggle', { actor_id: actor.id, list }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['actorList'] })
  })

  const aspect =
    web?.actorCardAspectRatio === '2:3' ? '2 / 3' : web?.actorCardAspectRatio === '9:16' ? '9 / 16' : '1 / 1'
  const fit = web?.actorCardScaleToFit ? 'contain' : 'cover'
  const opacity = actor.avail_count > 0 ? 1 : (web?.isAvailOpacity ?? 40) / 100

  return (
    <div className="group">
      <div
        className="relative cursor-pointer overflow-hidden rounded-xl bg-surface-3 ring-accent transition-all duration-150 group-hover:-translate-y-0.5 group-hover:shadow-xl group-hover:ring-2"
        style={{ aspectRatio: aspect }}
        onClick={onClick ?? (() => showActorDetails(actor.id))}
      >
        <img
          src={getImageURL(actor.image_url, '700x')}
          alt={actor.name}
          loading="lazy"
          onError={(e) => {
            const el = e.target as HTMLImageElement
            if (!el.src.endsWith('blank_female_profile.png'))
              el.src = `${import.meta.env.BASE_URL}blank_female_profile.png`
          }}
          className="absolute inset-0 h-full w-full"
          style={{ objectFit: fit, opacity }}
        />
        <div className="pointer-events-none absolute inset-x-0 bottom-0 flex flex-wrap justify-end gap-1 p-1.5">
          {actor.star_rating > 0 && (
            <span className="flex items-center gap-0.5 rounded-full bg-black/55 px-1.5 py-0.5 text-[10px] font-semibold text-white backdrop-blur-sm">
              ★ {actor.star_rating}
            </span>
          )}
          {actor.scene_rating_average !== undefined && Number(actor.scene_rating_average) > 0 && (
            <span
              title="Average scene rating"
              className="flex items-center gap-0.5 rounded-full bg-black/55 px-1.5 py-0.5 text-[10px] font-semibold text-white backdrop-blur-sm"
            >
              ∅ {Number(actor.scene_rating_average).toFixed(2)}
            </span>
          )}
        </div>
      </div>
      <div className="pt-1.5">
        <div className="truncate text-[13px] font-semibold">{actor.name}</div>
        <div className="flex items-center justify-between gap-1">
          <span className="flex gap-1">
            {web?.sceneFavourite && (
              <button
                onClick={() => toggle.mutate('favourite')}
                className={`rounded border px-1 py-0.5 text-[10px] ${actor.favourite ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
                title="Favourite"
              >
                ♥
              </button>
            )}
            {web?.sceneWatchlist && (
              <button
                onClick={() => toggle.mutate('watchlist')}
                className={`rounded border px-1 py-0.5 text-[10px] ${actor.watchlist ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
                title="Watchlist"
              >
                later
              </button>
            )}
          </span>
          <span className="text-[11px] text-muted">
            {actor.avail_count}/{actor.count}
            {actor.birth_date && !actor.birth_date.startsWith('0001-01-01') && ` · ${formatDate(actor.birth_date)}`}
          </span>
        </div>
      </div>
    </div>
  )
}

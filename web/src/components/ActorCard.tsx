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
  const web = state?.config?.web
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
        className="relative cursor-pointer overflow-hidden rounded-xl bg-surface-3 ring-accent transition-all duration-200 group-hover:-translate-y-1 group-hover:shadow-[0_10px_36px_rgba(0,0,0,0.45)] group-hover:ring-2"
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
        <div className="pointer-events-none absolute right-1.5 top-1.5 flex flex-wrap justify-end gap-1">
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
        <div className="scrim pointer-events-none absolute inset-x-0 bottom-0 p-2 pt-8">
          <div className="truncate text-[13px] font-semibold text-white drop-shadow">{actor.name}</div>
          <div className="text-[11px] text-white/75">
            {actor.avail_count}/{actor.count} scenes
          </div>
        </div>
      </div>
      <div className="flex items-center justify-between gap-1 pt-1.5">
        <span className="flex gap-1">
          {web?.sceneFavourite && (
            <button
              onClick={() => toggle.mutate('favourite')}
              className={`rounded-full border px-1.5 py-0.5 text-[10px] ${actor.favourite ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
              title="Favourite"
            >
              ♥
            </button>
          )}
          {web?.sceneWatchlist && (
            <button
              onClick={() => toggle.mutate('watchlist')}
              className={`rounded-full border px-1.5 py-0.5 text-[10px] ${actor.watchlist ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
              title="Watchlist"
            >
              later
            </button>
          )}
        </span>
        {actor.birth_date && !actor.birth_date.startsWith('0001-01-01') && (
          <span className="text-[11px] text-muted">{formatDate(actor.birth_date)}</span>
        )}
      </div>
    </div>
  )
}

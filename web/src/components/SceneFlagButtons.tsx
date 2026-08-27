import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Scene, WebOptions } from '../api/types'
import { patchSceneInCaches } from '../api/sceneCache'

type ToggleList =
  | 'watchlist'
  | 'trailerlist'
  | 'favourite'
  | 'needs_update'
  | 'watched'
  | 'is_hidden'
  | 'wishlist'

export function useSceneToggle() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ scene, list }: { scene: Scene; list: ToggleList }) =>
      api.post('/scene/toggle', { scene_id: scene.scene_id, list }),
    onMutate: ({ scene, list }) => {
      // optimistic toggle, mirroring the old UI
      const key =
        list === 'is_hidden'
          ? 'is_hidden'
          : list === 'needs_update'
            ? 'needs_update'
            : list === 'watched'
              ? 'is_watched'
              : list
      const current = scene[key as keyof Scene] as boolean
      patchSceneInCaches(queryClient, scene.id, { [key]: !current } as Partial<Scene>)
      return { key, current }
    },
    onError: (_error, { scene }, context) => {
      if (context) patchSceneInCaches(queryClient, scene.id, { [context.key]: context.current } as Partial<Scene>)
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: ['sceneList'], refetchType: 'none' })
  })
}

const BTN = 'rounded-full border px-2 py-0.5 text-[11px] font-medium leading-none transition-colors'

// The per-card / per-page scene flag buttons, gated by the Web UI options
// (Options → Web UI controls which buttons are visible). With `onTile` (used
// on SceneCard) the favourite/watchlist buttons are omitted — those are
// always-visible on-tile toggles instead.
export function SceneFlagButtons({
  scene,
  web,
  onEdit,
  onTile = false
}: {
  scene: Scene
  web?: WebOptions
  onEdit?: () => void
  onTile?: boolean
}) {
  const toggle = useSceneToggle()

  const t = (list: ToggleList) => toggle.mutate({ scene, list })

  return (
    <span className="inline-flex flex-wrap items-center gap-1">
      {web?.sceneHidden && (
        <button
          className={`${BTN} ${scene.is_hidden ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
          title={scene.is_hidden ? 'Unhide' : 'Hide'}
          onClick={(e) => { e.stopPropagation(); t('is_hidden') }}
        >
          {scene.is_hidden ? 'unhide' : 'hide'}
        </button>
      )}
      {/* watchlist + favourite live as on-tile toggles (SceneCard); keep the
          option-gated buttons only outside card context */}
      {!onTile && web?.sceneWatchlist && (
        <button
          className={`${BTN} ${scene.watchlist ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
          title="Watchlist"
          onClick={(e) => { e.stopPropagation(); t('watchlist') }}
        >
          watch later
        </button>
      )}
      {web?.sceneTrailerlist && (
        <button
          className={`${BTN} ${scene.trailerlist ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
          title="Trailer list"
          onClick={(e) => { e.stopPropagation(); t('trailerlist') }}
        >
          trailer
        </button>
      )}
      {!onTile && web?.sceneFavourite && (
        <button
          className={`${BTN} ${scene.favourite ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
          title="Favourite"
          onClick={(e) => { e.stopPropagation(); t('favourite') }}
        >
          ♥
        </button>
      )}
      {web?.sceneWishlist && !scene.is_available && (
        <button
          className={`${BTN} ${scene.wishlist ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
          title="Wishlist"
          onClick={(e) => { e.stopPropagation(); t('wishlist') }}
        >
          wish
        </button>
      )}
      {web?.sceneWatched && (
        <button
          className={`${BTN} ${scene.is_watched ? 'border-accent text-accent-strong' : 'border-line text-muted hover:text-fg'}`}
          title="Watched"
          onClick={(e) => { e.stopPropagation(); t('watched') }}
        >
          watched
        </button>
      )}
      {web?.sceneEdit && onEdit && (
        <button
          className={`${BTN} border-line text-muted hover:text-fg`}
          title="Edit"
          onClick={(e) => { e.stopPropagation(); onEdit() }}
        >
          edit
        </button>
      )}
    </span>
  )
}

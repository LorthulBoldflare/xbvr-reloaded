import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { AlternateSource, Scene } from '../api/types'
import { getImageURL, altSourceIconContext } from '../lib/image'
import { formatDate, safeHref } from '../lib/format'
import { useOptionsState } from '../api/hooks'
import { SceneFlagButtons, useSceneToggle } from './SceneFlagButtons'
import {
  BookmarkIcon, ClockIcon, CloudOffIcon, CuepointIcon, EyeIcon, FileIcon, GogglesIcon, HeartIcon, LinkIcon, PulseIcon, StarIcon, StorageOffIcon, SubtitlesIcon
} from './icons'

function Badge({ children, title }: { children: React.ReactNode; title?: string }) {
  return (
    <span
      title={title}
      className="flex items-center gap-0.5 rounded-full bg-black/55 px-1.5 py-0.5 text-[10px] font-semibold text-white backdrop-blur-sm"
    >
      {children}
    </span>
  )
}

// Always-visible on-tile toggle button (favourite heart / watchlist bookmark).
function TileToggle({
  children,
  title,
  active,
  activeClass,
  onClick
}: {
  children: React.ReactNode
  title: string
  active: boolean
  activeClass: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      title={title}
      onClick={(e) => {
        e.stopPropagation()
        onClick()
      }}
      className={`flex h-6 w-6 items-center justify-center rounded-full bg-black/55 backdrop-blur-sm transition-transform hover:scale-110 ${
        active ? activeClass : 'text-white/80'
      }`}
    >
      {children}
    </button>
  )
}

// One scene in the grid. Hover: accent border + preview video (when the
// scene has one). Click: navigate to the scene page (or call onOpen).
export function SceneCard({ scene, onOpen }: { scene: Scene; onOpen?: (scene: Scene) => void }) {
  const navigate = useNavigate()
  const { data: state } = useOptionsState()
  const web = state?.config?.web

  const [hover, setHover] = useState(false)
  const [altSources, setAltSources] = useState<AlternateSource[] | null>(null)
  const toggle = useSceneToggle()

  const files = scene.file ?? []
  const videoCount = files.filter((f) => f.type === 'video').length
  const scriptFiles = files.filter((f) => f.type === 'script')
  const hspCount = files.filter((f) => f.type === 'hsp').length
  const subCount = files.filter((f) => f.type === 'subtitles').length

  // Fixed 16:9 card frame; content fills the width, keeps its natural
  // proportions, and is vertically centered (overflow cropped).
  const opacity = scene.is_available ? 1 : (web?.isAvailOpacity ?? 40) / 100

  // Funscript heatmap strip(s): selected script first, or all
  const heatmaps = (() => {
    if (!web?.showScriptHeatmap) return []
    const withMap = scriptFiles.filter((f) => f.has_heatmap)
    if (web.showAllHeatmaps) return withMap
    const sel = withMap.find((f) => f.is_selected_script)
    return sel ? [sel] : withMap.slice(0, 1)
  })()

  // Alternate-source icons are fetched lazily on first hover (the old UI did
  // this per card on mount, which floods the API when scrolling).
  const loadAltSources = () => {
    if (altSources !== null) return
    api
      .get<AlternateSource[]>(`/scene/alternate_source/${scene.id}`, { toastOnError: false })
      .then((res) =>
        setAltSources(
          (res ?? []).filter(
            (a) => a.external_source.startsWith('alternate scene ') || a.external_source === 'stashdb scene'
          )
        )
      )
      .catch(() => setAltSources([]))
  }

  const altTitle = (a: AlternateSource): string => {
    try {
      const ext = JSON.parse(a.external_data)
      if (a.external_source.startsWith('alternate scene ')) return ext.scene?.title || 'No Title'
      return ext.title || 'No Title'
    } catch {
      return 'No Title'
    }
  }

  return (
    <div className="group">
      <div
        className="relative aspect-video cursor-pointer overflow-hidden rounded-xl bg-surface-3 ring-accent transition-all duration-200 group-hover:-translate-y-1 group-hover:shadow-[0_10px_36px_rgba(0,0,0,0.45)] group-hover:ring-2"
        onClick={() => (onOpen ? onOpen(scene) : navigate(`/scenes/${scene.scene_id}`))}
        onMouseEnter={() => {
          setHover(true)
          loadAltSources()
        }}
        onMouseLeave={() => setHover(false)}
      >
        <img
          src={getImageURL(scene.cover_url, '700x', scene.scene_id)}
          alt={scene.title}
          loading="lazy"
          onError={(e) => {
            const el = e.target as HTMLImageElement
            if (!el.src.endsWith('blank.png')) el.src = `${import.meta.env.BASE_URL}blank.png`
          }}
          className="absolute left-0 top-1/2 w-full -translate-y-1/2"
          style={{ opacity }}
        />
        {hover && scene.has_preview && (
          <video
            src={`/api/dms/preview/${scene.scene_id}`}
            autoPlay
            muted
            loop
            playsInline
            className="absolute left-0 top-1/2 w-full -translate-y-1/2"
          />
        )}

        {/* prominent "not available" indicator (the dimming alone is too subtle) */}
        {!scene.is_available && (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <span className="flex items-center gap-1.5 rounded-full bg-black/65 px-3 py-1.5 text-xs font-semibold text-white backdrop-blur-sm ring-1 ring-white/25">
              <CloudOffIcon className="h-4 w-4" />
              Not downloaded
            </span>
          </div>
        )}
        {scene.is_available && !scene.is_accessible && (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <span className="flex items-center gap-1.5 rounded-full bg-black/65 px-3 py-1.5 text-xs font-semibold text-warn backdrop-blur-sm ring-1 ring-warn/40">
              <StorageOffIcon className="h-4 w-4" />
              Storage offline
            </span>
          </div>
        )}

        {/* top-left: duration, then cuepoints */}
        <div className="pointer-events-none absolute left-1.5 top-1.5 flex gap-1">
          {scene.duration > 0 && (
            <Badge title="Duration">
              <ClockIcon /> {scene.duration}m
            </Badge>
          )}
          {(scene.cuepoints?.length ?? 0) > 0 && (
            <Badge title={`${scene.cuepoints!.length} cuepoints`}>
              <CuepointIcon /> {scene.cuepoints!.length}
            </Badge>
          )}
        </div>

        {/* top-right: info badges, watched, rating, watchlist + favourite toggles */}
        <div className="absolute right-1.5 top-1.5 flex items-center gap-1">
          {videoCount > 1 && !scene.is_multipart && (
            <Badge title={`${videoCount} video files`}>
              <FileIcon /> {videoCount}
            </Badge>
          )}
          {scene.is_scripted && (
            <Badge title="Scripted">
              <PulseIcon />
              {scriptFiles.length > 1 && <span>{scriptFiles.length}</span>}
            </Badge>
          )}
          {hspCount > 0 && web?.showHspFile && (
            <Badge title="HereSphere file">
              <GogglesIcon />
              {hspCount > 1 && <span>{hspCount}</span>}
            </Badge>
          )}
          {subCount > 0 && web?.showSubtitlesFile && (
            <Badge title="Subtitles">
              <SubtitlesIcon />
              {subCount > 1 && <span>{subCount}</span>}
            </Badge>
          )}
          {scene.is_watched && (
            <Badge title="Watched">
              <EyeIcon />
            </Badge>
          )}
          {scene.star_rating > 0 && (
            <Badge title={`Rating ${scene.star_rating}`}>
              <StarIcon className="h-3.5 w-3.5 text-warn" /> {scene.star_rating}
            </Badge>
          )}
          <TileToggle
            title={scene.watchlist ? 'Remove from watchlist' : 'Add to watchlist'}
            active={scene.watchlist}
            activeClass="text-ok"
            onClick={() => toggle.mutate({ scene, list: 'watchlist' })}
          >
            <BookmarkIcon filled={scene.watchlist} />
          </TileToggle>
          <TileToggle
            title={scene.favourite ? 'Unfavourite' : 'Favourite'}
            active={scene.favourite}
            activeClass="text-danger"
            onClick={() => toggle.mutate({ scene, list: 'favourite' })}
          >
            <HeartIcon filled={scene.favourite} />
          </TileToggle>
        </div>

        {/* gradient scrim with title + meta overlaid on the artwork */}
        <div className="scrim pointer-events-none absolute inset-x-0 bottom-0 flex flex-col justify-end p-2.5 pt-10">
          {heatmaps.length > 0 && (
            <div className="mb-1.5 space-y-0.5">
              {heatmaps.map((f) => (
                <img
                  key={f.id}
                  src={`/api/dms/heatmap/${f.id}`}
                  alt="funscript heatmap"
                  className="h-2.5 w-full rounded-full opacity-90"
                  loading="lazy"
                />
              ))}
            </div>
          )}
          <div className="line-clamp-2 text-[13px] font-semibold leading-snug text-white drop-shadow">
            {scene.title}
          </div>
          <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-white/75">
            <span
              className={scene.is_subscribed ? 'rounded bg-accent px-1 font-semibold text-accent-fg' : ''}
            >
              {scene.site}
            </span>
            <span>{formatDate(scene.release_date)}</span>
          </div>
        </div>
      </div>

      {/* actions row */}
      <div className="flex items-start justify-between gap-1 pt-1.5">
        <div className="min-w-0">
          <SceneFlagButtons scene={scene} onTile onEdit={() => navigate(`/scenes/${scene.scene_id}?edit=1`)} />
        </div>
        <div className="flex shrink-0 items-center gap-1 pt-0.5">
          {scene.members_url && (
            <a
              href={safeHref(scene.members_url)}
              target="_blank"
              rel="noreferrer"
              title="Members link"
              className="text-muted hover:text-accent"
              onClick={(e) => e.stopPropagation()}
            >
              <LinkIcon />
            </a>
          )}
          <a
            href={safeHref(scene.scene_url)}
            target="_blank"
            rel="noreferrer"
            title="Open scene page on source site"
            onClick={(e) => e.stopPropagation()}
            className="text-muted hover:text-accent"
          >
            <LinkIcon />
          </a>
        </div>
      </div>
      {altSources && altSources.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {altSources.map((a, i) => (
            <a key={i} href={safeHref(a.url)} target="_blank" rel="noreferrer" title={altTitle(a)}>
              <img
                src={getImageURL(a.site_icon, '20x', altSourceIconContext(a))}
                alt=""
                className="h-5 w-5 rounded"
                loading="lazy"
                onError={(e) => ((e.target as HTMLImageElement).style.display = 'none')}
              />
            </a>
          ))}
        </div>
      )}
    </div>
  )
}

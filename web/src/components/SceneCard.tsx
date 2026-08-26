import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { AlternateSource, Scene } from '../api/types'
import { getImageURL } from '../lib/image'
import { formatDate, safeHref } from '../lib/format'
import { useOptionsState } from '../api/hooks'
import { SceneFlagButtons } from './SceneFlagButtons'
import {
  ClockIcon, CuepointIcon, EyeIcon, FileIcon, GogglesIcon, LinkIcon, PulseIcon, StarIcon, SubtitlesIcon
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

// One scene in the grid. Hover: accent border + preview video (when the
// scene has one). Click: navigate to the scene page (or call onOpen).
export function SceneCard({ scene, onOpen }: { scene: Scene; onOpen?: (scene: Scene) => void }) {
  const navigate = useNavigate()
  const { data: state } = useOptionsState()
  const web = state?.currentState?.web

  const [hover, setHover] = useState(false)
  const [altSources, setAltSources] = useState<AlternateSource[] | null>(null)

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
        className="relative aspect-video cursor-pointer overflow-hidden rounded-xl bg-surface-3 ring-accent transition-all duration-150 group-hover:-translate-y-0.5 group-hover:shadow-xl group-hover:ring-2"
        onClick={() => (onOpen ? onOpen(scene) : navigate(`/scenes/${scene.scene_id}`))}
        onMouseEnter={() => {
          setHover(true)
          loadAltSources()
        }}
        onMouseLeave={() => setHover(false)}
      >
        <img
          src={getImageURL(scene.cover_url)}
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
        <div className="pointer-events-none absolute inset-x-0 bottom-0 flex flex-col items-end gap-1 p-1.5">
          <div className="flex flex-wrap justify-end gap-1">
            {scene.is_watched && !web?.sceneWatched && (
              <Badge title="Watched">
                <EyeIcon />
              </Badge>
            )}
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
            {(scene.cuepoints?.length ?? 0) > 0 && web?.sceneCuepoint && (
              <Badge title="Cuepoints">
                <CuepointIcon />
                {(scene.cuepoints?.length ?? 0) > 1 && <span>{scene.cuepoints!.length}</span>}
              </Badge>
            )}
            {scene.star_rating > 0 && (
              <Badge title={`Rating ${scene.star_rating}`}>
                <StarIcon className="h-3.5 w-3.5 text-warn" /> {scene.star_rating}
              </Badge>
            )}
            {scene.duration > 0 && web?.sceneDuration && (
              <Badge title="Duration">
                <ClockIcon /> {scene.duration}m
              </Badge>
            )}
          </div>
          {heatmaps.length > 0 && (
            <div className="w-full space-y-0.5">
              {heatmaps.map((f) => (
                <img
                  key={f.id}
                  src={`/api/dms/heatmap/${f.id}`}
                  alt="funscript heatmap"
                  className="h-3.5 w-full rounded-full border border-line-strong object-fill"
                  loading="lazy"
                />
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="pt-1.5">
        <div className="line-clamp-2 min-h-[2.6em] text-[13px] font-semibold leading-snug">{scene.title}</div>
        <div className="mt-0.5 flex items-start justify-between gap-1">
          <div className="min-w-0">
            <SceneFlagButtons scene={scene} onEdit={() => navigate(`/scenes/${scene.scene_id}?edit=1`)} />
          </div>
          <div className="shrink-0 text-right text-[11px] leading-tight text-muted">
            <div className="flex items-center justify-end gap-1">
              {scene.members_url && (
                <a
                  href={safeHref(scene.members_url)}
                  target="_blank"
                  rel="noreferrer"
                  title="Members link"
                  className="hover:text-accent"
                  onClick={(e) => e.stopPropagation()}
                >
                  <LinkIcon />
                </a>
              )}
              <a
                href={safeHref(scene.scene_url)}
                target="_blank"
                rel="noreferrer"
                onClick={(e) => e.stopPropagation()}
                className={`rounded px-1 font-semibold hover:bg-accent-soft hover:text-accent-strong ${
                  scene.is_subscribed ? 'bg-accent-soft text-accent-strong' : ''
                }`}
              >
                {scene.site}
              </a>
            </div>
            <div>{formatDate(scene.release_date)}</div>
          </div>
        </div>
        {altSources && altSources.length > 0 && (
          <div className="mt-1 flex flex-wrap gap-1">
            {altSources.map((a, i) => (
              <a key={i} href={safeHref(a.url)} target="_blank" rel="noreferrer" title={altTitle(a)}>
                <img
                  src={getImageURL(a.site_icon, '20x')}
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
    </div>
  )
}

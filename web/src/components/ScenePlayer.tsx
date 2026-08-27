import { useEffect, useRef, useState } from 'react'
import type { File } from '../api/types'
import { getImageURL } from '../lib/image'

/*
 * ScenePlayer — flat <video> playback of a scene file.
 *
 * REPLACEABLE SEAM: this component owns the actual media element. Callers
 * only pass a file + optional poster and receive time updates. To add a
 * WebXR/VR player later, swap the internals of this component (keeping the
 * props interface) — no caller changes needed.
 */
export function ScenePlayer({
  file,
  poster,
  posterContext = 'scene-0',
  onTimeUpdate,
  className = ''
}: {
  file: File | null
  poster?: string
  // image proxy context for the poster, e.g. sceneContext(scene.scene_id)
  posterContext?: string
  className?: string
  // Called with the current playback position (seconds); used by the
  // cuepoint editor's "current time" button.
  onTimeUpdate?: (t: number) => void
}) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [lastSkip, setLastSkip] = useState<number | null>(null)

  useEffect(() => {
    // Load a new source when the file changes.
    videoRef.current?.load()
  }, [file?.id])

  useEffect(() => {
    const v = videoRef.current
    if (!v || !onTimeUpdate) return
    const cb = () => onTimeUpdate(v.currentTime)
    v.addEventListener('timeupdate', cb)
    return () => v.removeEventListener('timeupdate', cb)
  }, [onTimeUpdate])

  const skip = (secs: number) => {
    const v = videoRef.current
    if (!v) return
    v.currentTime = Math.max(0, Math.min(v.duration || Infinity, v.currentTime + secs))
    setLastSkip(secs)
  }

  if (!file) return null

  const SKIPS = [-300, -120, -60, -30, -10, -5, 5, 10, 30, 60, 120, 300]

  return (
    <div className={className}>
      <video
        ref={videoRef}
        controls
        playsInline
        poster={poster ? getImageURL(poster, '700,fit', posterContext) : undefined}
        className="w-full rounded-xl bg-black"
      >
        <source src={`/api/dms/file/${file.id}?dnt=true`} />
      </video>
      <div className="mt-1.5 flex flex-wrap justify-center gap-1">
        {SKIPS.map((s) => (
          <button
            key={s}
            onClick={() => skip(s)}
            className={`rounded px-1.5 py-0.5 font-mono text-[11px] ${
              lastSkip === s ? 'bg-accent text-white' : 'bg-surface-3 text-muted hover:text-fg'
            }`}
          >
            {s > 0 ? `+${s}` : s}
          </button>
        ))}
      </div>
    </div>
  )
}

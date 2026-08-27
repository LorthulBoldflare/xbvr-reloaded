import { useEffect, useMemo, useRef, useState } from 'react'
import type { SceneImage } from '../../api/types'
import { getImageURL, sceneContext } from '../../lib/image'

// Image carousel for the scene page. Fixed 16:9 frame; the image fills the
// height and keeps its ratio (width follows), centered over a blurred copy of
// itself so portrait images don't look jarring. The scene's cover/thumbnail
// image is excluded (it's already used as the player poster / card cover).
export function SceneGallery({ images, coverUrl, sceneId }: { images: SceneImage[]; coverUrl?: string; sceneId: string }) {
  const gallery = useMemo(
    () => images.filter((img) => img.type !== 'cover' && img.url !== coverUrl),
    [images, coverUrl]
  )

  const [index, setIndex] = useState(0)
  const stripRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (index >= gallery.length) setIndex(0)
  }, [gallery.length, index])

  useEffect(() => {
    stripRef.current?.querySelector(`[data-i="${index}"]`)?.scrollIntoView({ inline: 'center', block: 'nearest' })
  }, [index])

  const current = gallery[index]
  const go = (d: number) => setIndex((i) => (i + d + gallery.length) % gallery.length)

  // Arrow keys are handled by the scene page's global handler via this event.
  useEffect(() => {
    const onNav = (e: Event) => go((e as CustomEvent<number>).detail)
    window.addEventListener('scene-gallery-nav', onNav)
    return () => window.removeEventListener('scene-gallery-nav', onNav)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gallery.length])

  if (gallery.length === 0) return null

  return (
    <div>
      <div className="relative aspect-video overflow-hidden rounded-xl bg-surface-3">
        {/* blurred backdrop filling the letterbox area */}
        <img
          key={`bg-${current.url}`}
          src={getImageURL(current.url, '700x', sceneContext(sceneId))}
          alt=""
          aria-hidden
          className="absolute inset-0 h-full w-full scale-110 object-cover opacity-40 blur-xl"
        />
        <img
          key={current.url}
          src={getImageURL(current.url, '700,fit', sceneContext(sceneId))}
          alt=""
          onError={(e) => {
            const el = e.target as HTMLImageElement
            if (!el.src.endsWith('blank.png')) el.src = `${import.meta.env.BASE_URL}blank.png`
          }}
          className="relative mx-auto block h-full w-auto max-w-full"
        />
        {gallery.length > 1 && (
          <>
            <button
              onClick={() => go(-1)}
              className="absolute left-2 top-1/2 -translate-y-1/2 rounded-full bg-black/50 px-2.5 py-1.5 text-white hover:bg-black/70"
              aria-label="Previous image"
            >
              ←
            </button>
            <button
              onClick={() => go(1)}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full bg-black/50 px-2.5 py-1.5 text-white hover:bg-black/70"
              aria-label="Next image"
            >
              →
            </button>
          </>
        )}
      </div>
      {gallery.length > 1 && (
        <div ref={stripRef} className="mt-2 flex gap-1 overflow-x-auto pb-1">
          {gallery.map((img, i) => (
            <button key={i} data-i={i} onClick={() => setIndex(i)} className="shrink-0">
              <img
                src={getImageURL(img.url, 'x40', sceneContext(sceneId))}
                alt=""
                className={`h-10 rounded object-cover ${i === index ? 'ring-2 ring-accent' : 'opacity-60 hover:opacity-100'}`}
              />
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export function parseSceneImages(json: string): SceneImage[] {
  try {
    const arr = JSON.parse(json || '[]') as SceneImage[]
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

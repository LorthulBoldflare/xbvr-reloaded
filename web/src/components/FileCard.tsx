import { useNavigate } from 'react-router-dom'
import type { File } from '../api/types'
import { formatDate, prettyBytes } from '../lib/format'
import { FileIcon } from './icons'

// Card for an unmatched file inside the scene grid — same shape/hover as
// SceneCard, but no scene metadata and no preview video.
export function FileCard({ file }: { file: File }) {
  const navigate = useNavigate()
  return (
    <div className="group">
      <div
        className="relative flex aspect-video cursor-pointer items-center justify-center overflow-hidden rounded-xl bg-surface-3 ring-accent transition-all duration-150 group-hover:-translate-y-0.5 group-hover:shadow-xl group-hover:ring-2"
        onClick={() => navigate(`/files/${file.id}`)}
        title={`${file.path}/${file.filename}`}
      >
        <FileIcon className="h-12 w-12 text-muted" />
        <span className="absolute left-1.5 top-1.5 rounded-full bg-warn/90 px-2 py-0.5 text-[10px] font-bold uppercase text-black">
          unmatched
        </span>
        <div className="pointer-events-none absolute inset-x-0 bottom-0 flex flex-wrap justify-end gap-1 p-1.5 text-[10px]">
          <span className="rounded-full bg-black/55 px-1.5 py-0.5 font-semibold text-white backdrop-blur-sm">
            {prettyBytes(file.size)}
          </span>
          {file.video_width > 0 && (
            <span className="rounded-full bg-black/55 px-1.5 py-0.5 font-semibold text-white backdrop-blur-sm">
              {file.video_width}×{file.video_height}
            </span>
          )}
        </div>
      </div>
      <div className="pt-1.5">
        <div className="line-clamp-2 min-h-[2.6em] break-all text-[13px] font-semibold leading-snug">
          {file.filename}
        </div>
        <div className="text-right text-[11px] text-muted">{formatDate(file.created_time)}</div>
      </div>
    </div>
  )
}

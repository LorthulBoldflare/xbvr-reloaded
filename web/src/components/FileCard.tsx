import { memo } from 'react'
import { useNavigate } from 'react-router-dom'
import type { File } from '../api/types'
import { formatDate, prettyBytes } from '../lib/format'
import { FileIcon } from './icons'

// Card for an unmatched file inside the scene grid — same tile shape/hover
// as SceneCard, but no scene metadata and no preview video.
export const FileCard = memo(function FileCard({ file }: { file: File }) {
  const navigate = useNavigate()
  return (
    <div className="media-grid-item group">
      <div
        className="media-card relative aspect-video cursor-pointer overflow-hidden rounded-xl bg-surface-2 ring-warn transition-transform duration-200 group-hover:-translate-y-1 group-hover:ring-2"
        onClick={() => navigate(`/files/${file.id}`)}
        title={`${file.path}/${file.filename}`}
      >
        <FileIcon className="absolute left-1/2 top-1/2 h-10 w-10 -translate-x-1/2 -translate-y-1/2 text-muted" />
        <div className="scrim pointer-events-none absolute inset-x-0 bottom-0 flex flex-col justify-end p-2.5 pt-8">
          <div className="line-clamp-2 break-all text-[12px] font-medium leading-snug text-white drop-shadow">
            {file.filename}
          </div>
          <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-white/75">
            <span className="rounded bg-warn px-1 text-[10px] font-bold uppercase text-black">unmatched</span>
            <span>{prettyBytes(file.size)}</span>
            {file.video_width > 0 && (
              <span>
                {file.video_width}×{file.video_height}
              </span>
            )}
            <span>{formatDate(file.created_time)}</span>
          </div>
        </div>
      </div>
    </div>
  )
})

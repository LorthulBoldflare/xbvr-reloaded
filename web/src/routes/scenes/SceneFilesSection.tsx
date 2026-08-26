import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { File, Scene } from '../../api/types'
import { formatDate, humanizeSeconds, prettyBytes } from '../../lib/format'
import { useUIStore } from '../../store/ui'
import { GogglesIcon, PlayIcon, PulseIcon, SubtitlesIcon, TrashIcon } from '../../components/icons'

// Files section of the scene page: video/script/subtitle/hsp rows with
// metadata; play and select-script always available; unmatch/delete only in
// edit mode.
export function SceneFilesSection({
  scene,
  editMode,
  onPlay
}: {
  scene: Scene
  editMode: boolean
  onPlay: (file: File) => void
}) {
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const files = [...(scene.file ?? [])].sort((a, b) => (a.type === 'video' ? -1 : b.type === 'video' ? 1 : 0))

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['scene'] })
    queryClient.invalidateQueries({ queryKey: ['sceneList'] })
  }

  const selectScript = useMutation({
    mutationFn: (fileId: number) => api.post(`/scene/selectscript/${scene.id}`, { file_id: fileId }),
    onSuccess: invalidate
  })
  const unmatch = useMutation({
    mutationFn: (fileId: number) => api.post('/files/unmatch', { file_id: fileId }),
    onSuccess: invalidate
  })
  const deleteFile = useMutation({
    mutationFn: (fileId: number) => api.delete(`/files/file/${fileId}`),
    onSuccess: invalidate
  })

  if (files.length === 0) {
    return <div className="text-sm text-muted">No files matched to this scene.</div>
  }

  return (
    <div className="space-y-1">
      {files.map((f) => (
        <div key={f.id} className="flex flex-wrap items-center gap-2 rounded-lg border border-line bg-surface-2 px-3 py-2">
          <span className="flex items-center gap-1 text-xs font-semibold uppercase text-muted">
            {f.type === 'video' && <PlayIcon />}
            {f.type === 'script' && <PulseIcon />}
            {f.type === 'hsp' && <GogglesIcon />}
            {f.type === 'subtitles' && <SubtitlesIcon />}
            {f.type}
          </span>
          <span className="min-w-0 flex-1">
            <span className="block truncate font-mono text-xs">{f.filename}</span>
            <span className="block truncate text-[11px] text-muted">{f.path}</span>
          </span>
          <span className="text-right text-[11px] text-muted">
            {prettyBytes(f.size)}
            {f.video_bitrate > 0 && ` @ ${(f.video_bitrate / 1e6).toFixed(1)} Mb/s`}
            {f.video_width > 0 && ` · ${f.video_width}×${f.video_height} ${f.video_codec_name}`}
            {f.projection && ` · ${f.projection}`}
            {f.duration > 0 && ` · ${humanizeSeconds(f.duration)}`}
            {` · ${formatDate(f.created_time)}`}
          </span>
          {f.type === 'video' && (
            <button
              onClick={() => onPlay(f)}
              className="rounded-lg bg-accent px-2.5 py-1 text-xs font-semibold text-white"
            >
              Play
            </button>
          )}
          {f.type === 'script' && (
            <button
              onClick={() => selectScript.mutate(f.id)}
              title="Select as the script for this scene"
              className={`rounded-lg border px-2 py-1 text-xs ${
                f.is_selected_script ? 'border-ok bg-ok/10 text-ok' : 'border-line text-muted hover:text-fg'
              }`}
            >
              {f.is_selected_script ? 'selected' : 'select script'}
            </button>
          )}
          {f.has_heatmap && (
            <img src={`/api/dms/heatmap/${f.id}`} alt="heatmap" className="h-4 w-24 rounded-full border border-line-strong" />
          )}
          {editMode && (
            <>
              <button
                onClick={async () => {
                  if (await askConfirm({ title: `Unmatch ${f.filename}?` })) unmatch.mutate(f.id)
                }}
                className="rounded-lg border border-line px-2 py-1 text-xs text-muted hover:text-fg"
              >
                unmatch
              </button>
              <button
                onClick={async () => {
                  if (
                    await askConfirm({
                      title: `Delete ${f.filename} from disk?`,
                      message: 'This removes the file from storage. This cannot be undone.',
                      danger: true
                    })
                  )
                    deleteFile.mutate(f.id)
                }}
                className="rounded-lg border border-line px-2 py-1 text-xs text-muted hover:text-danger"
                title="Delete from disk"
              >
                <TrashIcon />
              </button>
            </>
          )}
        </div>
      ))}
    </div>
  )
}

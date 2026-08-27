import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { File, Scene } from '../../api/types'
import { formatDate, humanizeSeconds, prettyBytes } from '../../lib/format'
import { useUIStore } from '../../store/ui'
import { useToastStore } from '../../store/toasts'
import { ScenePlayer } from '../../components/ScenePlayer'
import { MatchSceneModal } from '../../components/MatchSceneModal'
import { CreateSceneModal } from '../../components/CreateSceneModal'
import { TrashIcon } from '../../components/icons'

// File page: for unmatched files this shows NO scene information — just the
// player, file metadata, and a "Match to scene" button.
export function FilePage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const toast = useToastStore.getState()
  const [matchOpen, setMatchOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)

  const { data: file, isLoading, isError } = useQuery({
    queryKey: ['file', id],
    queryFn: () => api.get<File>(`/files/file/${id}`, { toastOnError: false }),
    retry: false
  })

  // When the file is matched, resolve the owning scene (numeric PK works).
  const { data: scene } = useQuery({
    queryKey: ['scene', file?.scene_id],
    queryFn: () => api.get<Scene>(`/scene/${file!.scene_id}`, { toastOnError: false }),
    enabled: !!file && file.scene_id > 0
  })

  const deleteFile = useMutation({
    mutationFn: () => api.delete(`/files/file/${file!.id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['files'] })
      navigate('/')
    }
  })

  const unmatch = useMutation({
    mutationFn: () => api.post('/files/unmatch', { file_id: file!.id }),
    onSuccess: () => {
      queryClient.invalidateQueries()
      toast.success('File unmatched')
    }
  })

  if (isLoading) return <div className="py-16 text-center text-muted">Loading…</div>
  if (isError || !file || !file.id) {
    toast.error('File not found')
    navigate('/', { replace: true })
    return null
  }

  const matched = file.scene_id > 0
  const isVideo = file.type === 'video'

  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <h1 className="min-w-0 flex-1 truncate font-mono text-lg">{file.filename}</h1>
        {matched ? (
          <>
            {scene && scene.id !== 0 && (
              <Link
                to={`/scenes/${scene.scene_id}`}
                className="rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white"
              >
                View scene: {scene.title}
              </Link>
            )}
            <button
              onClick={async () => {
                if (await askConfirm({ title: 'Unmatch this file from its scene?' })) unmatch.mutate()
              }}
              className="rounded-lg border border-line px-3 py-1.5 text-sm text-muted hover:text-fg"
            >
              Unmatch
            </button>
          </>
        ) : (
          <>
            <span className="rounded-full bg-warn/20 px-2 py-0.5 text-xs font-semibold uppercase text-warn">
              unmatched
            </span>
            <button
              onClick={() => setMatchOpen(true)}
              className="rounded-lg bg-accent px-4 py-1.5 text-sm font-semibold text-white"
            >
              Match to scene
            </button>
            <button
              onClick={() => setCreateOpen(true)}
              className="rounded-lg border border-line px-3 py-1.5 text-sm hover:bg-surface-2"
            >
              Create custom scene
            </button>
          </>
        )}
        <button
          onClick={async () => {
            if (
              await askConfirm({
                title: `Delete ${file.filename} from disk?`,
                message: 'This removes the file from storage. This cannot be undone.',
                danger: true
              })
            )
              deleteFile.mutate()
          }}
          className="rounded-lg border border-line px-2.5 py-1.5 text-sm text-muted hover:text-danger"
          title="Delete from disk"
        >
          <TrashIcon />
        </button>
      </div>

      {isVideo && <ScenePlayer file={file} className="mb-4" />}

      <table className="w-full text-sm">
        <tbody>
          {[
            ['Path', file.path],
            ['Size', prettyBytes(file.size)],
            file.video_width > 0 && ['Resolution', `${file.video_width}×${file.video_height}`],
            file.video_codec_name && ['Codec', file.video_codec_name],
            file.video_bitrate > 0 && ['Bitrate', `${(file.video_bitrate / 1e6).toFixed(1)} Mb/s`],
            file.video_avgfps_val > 0 && ['Framerate', `${file.video_avgfps_val.toFixed(0)} fps`],
            file.projection && ['Projection', file.projection],
            file.duration > 0 && ['Duration', humanizeSeconds(file.duration)],
            ['Created', formatDate(file.created_time)]
          ]
            .filter(Boolean)
            .map((row) => {
              const [k, v] = row as [string, string]
              return (
                <tr key={k} className="border-t border-line">
                  <td className="w-32 py-1.5 pr-2 text-xs font-semibold uppercase text-muted">{k}</td>
                  <td className="py-1.5">{v}</td>
                </tr>
              )
            })}
        </tbody>
      </table>

      <MatchSceneModal
        file={file}
        open={matchOpen}
        onClose={() => setMatchOpen(false)}
        onMatched={() => {
          queryClient.invalidateQueries()
          toast.success('File matched')
        }}
      />
      <CreateSceneModal
        file={file}
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onDone={() => queryClient.invalidateQueries()}
      />
    </div>
  )
}

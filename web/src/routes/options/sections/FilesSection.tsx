import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { File } from '../../../api/types'
import { formatDate, humanizeSeconds, prettyBytes } from '../../../lib/format'
import { useUIStore } from '../../../store/ui'
import { MatchSceneModal } from '../../../components/MatchSceneModal'
import { CreateSceneModal } from '../../../components/CreateSceneModal'
import { SectionCard, btnCls, inputCls } from '../common'

const RESOLUTIONS = ['below4k', '4k', '5k', '6k', 'above6k']
const BITRATES = ['low', 'medium', 'high', 'ultra']
const FRAMERATES = ['30fps', '60fps', 'other']
const SORTABLE: [string, string][] = [
  ['created_time', 'Created'],
  ['filename', 'File'],
  ['size', 'Size'],
  ['video_bitrate', 'Bitrate'],
  ['duration', 'Duration']
]

// The old Files page, relocated into Options.
export function FilesSection() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)

  const [state, setState] = useState<'all' | 'matched' | 'unmatched'>('unmatched')
  const [filename, setFilename] = useState('')
  const [createdFrom, setCreatedFrom] = useState('')
  const [createdTo, setCreatedTo] = useState('')
  const [resolutions, setResolutions] = useState<string[]>([])
  const [bitrates, setBitrates] = useState<string[]>([])
  const [framerates, setFramerates] = useState<string[]>([])
  const [sort, setSort] = useState('created_time_desc')
  const [page, setPage] = useState(0)
  const [matchFile, setMatchFile] = useState<File | null>(null)
  const [createFile, setCreateFile] = useState<File | null>(null)

  const body = useMemo(
    () => ({
      state: state === 'all' ? '' : state,
      filename,
      createdDate: createdFrom && createdTo ? [createdFrom, createdTo] : [],
      resolutions,
      framerates,
      bitrates,
      sort
    }),
    [state, filename, createdFrom, createdTo, resolutions, framerates, bitrates, sort]
  )

  const { data: files } = useQuery({
    queryKey: ['files', body],
    queryFn: () => api.post<File[]>('/files/list', body)
  })

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['files'] })
    queryClient.invalidateQueries({ queryKey: ['unmatchedFiles'] })
  }

  const unmatch = useMutation({
    mutationFn: (fileId: number) => api.post('/files/unmatch', { file_id: fileId }),
    onSuccess: invalidate
  })
  const deleteFile = useMutation({
    mutationFn: (fileId: number) => api.delete(`/files/file/${fileId}`),
    onSuccess: invalidate
  })

  const all = files ?? []
  const pageSize = 20
  const pageCount = Math.max(1, Math.ceil(all.length / pageSize))
  const rows = all.slice(page * pageSize, (page + 1) * pageSize)

  const toggleSort = (field: string) =>
    setSort((s) => (s === `${field}_desc` ? `${field}_asc` : `${field}_desc`))

  const sortIndicator = (field: string) =>
    sort === `${field}_desc` ? ' ↓' : sort === `${field}_asc` ? ' ↑' : ''

  const multi = (values: string[], set: (v: string[]) => void, options: string[]) => (
    <div className="flex flex-wrap gap-1">
      {options.map((o) => (
        <button
          key={o}
          onClick={() => set(values.includes(o) ? values.filter((x) => x !== o) : [...values, o])}
          className={`rounded-full border px-2 py-0.5 text-xs ${
            values.includes(o) ? 'border-accent bg-accent-soft text-accent-strong' : 'border-line text-muted'
          }`}
        >
          {o}
        </button>
      ))}
    </div>
  )

  return (
    <SectionCard title={`Files (${all.length})`}>
      <div className="mb-3 flex flex-wrap items-end gap-3">
        <div className="flex gap-1">
          {(['all', 'matched', 'unmatched'] as const).map((s) => (
            <button
              key={s}
              onClick={() => {
                setState(s)
                setPage(0)
              }}
              className={`rounded-lg px-2.5 py-1 text-xs font-medium ${
                state === s ? 'bg-accent-soft text-accent-strong' : 'text-muted hover:bg-surface-2'
              }`}
            >
              {s[0].toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>
        <input
          value={filename}
          onChange={(e) => {
            setFilename(e.target.value)
            setPage(0)
          }}
          placeholder="Filename filter"
          className={`${inputCls} w-48`}
        />
        <label className="text-xs text-muted">
          Created from
          <input type="date" value={createdFrom} onChange={(e) => setCreatedFrom(e.target.value)} className={`${inputCls} mt-0.5`} />
        </label>
        <label className="text-xs text-muted">
          to
          <input type="date" value={createdTo} onChange={(e) => setCreatedTo(e.target.value)} className={`${inputCls} mt-0.5`} />
        </label>
        <div>{multi(resolutions, setResolutions, RESOLUTIONS)}</div>
        <div>{multi(bitrates, setBitrates, BITRATES)}</div>
        <div>{multi(framerates, setFramerates, FRAMERATES)}</div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-xs uppercase text-muted">
              {SORTABLE.map(([f, label]) => (
                <th key={f} className="cursor-pointer py-1 pr-2 hover:text-fg" onClick={() => toggleSort(f)}>
                  {label}
                  {sortIndicator(f)}
                </th>
              ))}
              <th className="py-1 pr-2">Resolution</th>
              <th className="py-1 pr-2">FPS</th>
              <th className="py-1">Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((f) => (
              <tr key={f.id} className="border-t border-line">
                <td className="py-1.5 pr-2">
                  <div className="max-w-72 truncate font-mono text-xs" title={`${f.path}/${f.filename}`}>
                    {f.filename}
                  </div>
                  <div className="max-w-72 truncate text-[10px] text-muted">{f.path}</div>
                </td>
                <td className="py-1.5 pr-2 text-xs">{formatDate(f.created_time)}</td>
                <td className="py-1.5 pr-2 text-xs">{prettyBytes(f.size)}</td>
                <td className="py-1.5 pr-2 text-xs">{f.video_bitrate ? `${(f.video_bitrate / 1e6).toFixed(1)} Mb/s` : ''}</td>
                <td className="py-1.5 pr-2 text-xs">{f.duration ? humanizeSeconds(f.duration) : ''}</td>
                <td className="py-1.5 pr-2 text-xs">{f.video_width ? `${f.video_width}×${f.video_height}` : ''}</td>
                <td className="py-1.5 pr-2 text-xs">{f.video_avgfps_val ? Math.round(f.video_avgfps_val) : ''}</td>
                <td className="py-1.5">
                  <div className="flex gap-1">
                    {f.type === 'video' && (
                      <button onClick={() => navigate(`/files/${f.id}`)} className={`${btnCls} px-2 py-0.5 text-xs`}>
                        Play
                      </button>
                    )}
                    {f.scene_id === 0 ? (
                      <button onClick={() => setMatchFile(f)} className={`${btnCls} px-2 py-0.5 text-xs`}>
                        Match
                      </button>
                    ) : (
                      <button
                        onClick={async () => {
                          if (await askConfirm({ title: `Unmatch ${f.filename}?` })) unmatch.mutate(f.id)
                        }}
                        className={`${btnCls} px-2 py-0.5 text-xs`}
                      >
                        Unmatch
                      </button>
                    )}
                    {f.scene_id === 0 && (
                      <button onClick={() => setCreateFile(f)} className={`${btnCls} px-2 py-0.5 text-xs`}>
                        + Scene
                      </button>
                    )}
                    <button
                      onClick={async () => {
                        if (
                          await askConfirm({
                            title: `Delete ${f.filename} from disk?`,
                            message: 'This cannot be undone.',
                            danger: true
                          })
                        )
                          deleteFile.mutate(f.id)
                      }}
                      className={`${btnCls} px-2 py-0.5 text-xs text-danger`}
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {rows.length === 0 && <div className="py-8 text-center text-sm text-muted">No files match</div>}
      </div>

      <div className="mt-3 flex items-center justify-center gap-2 text-sm">
        <button disabled={page === 0} onClick={() => setPage((p) => p - 1)} className={`${btnCls} px-2 py-1`}>
          ←
        </button>
        <span className="text-muted">
          {page + 1} / {pageCount}
        </span>
        <button disabled={page >= pageCount - 1} onClick={() => setPage((p) => p + 1)} className={`${btnCls} px-2 py-1`}>
          →
        </button>
      </div>

      <MatchSceneModal file={matchFile} open={matchFile !== null} onClose={() => setMatchFile(null)} onMatched={invalidate} />
      <CreateSceneModal file={createFile} open={createFile !== null} onClose={() => setCreateFile(null)} onDone={invalidate} />
    </SectionCard>
  )
}

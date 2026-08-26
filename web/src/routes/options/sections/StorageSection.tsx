import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { Volume, WebhooksConfig } from '../../../api/types'
import { formatDate, prettyBytes } from '../../../lib/format'
import { useOptionsStorage } from '../../../api/hooks'
import { useUIStore } from '../../../store/ui'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { ListEditor } from '../../../components/ListEditor'
import { SectionCard, Field, SaveButton, btnCls, btnDangerCls, inputCls } from '../common'

// Storage: volumes, add local/cloud, webhooks, matching options.
export function StorageSection() {
  const { data: storage } = useOptionsStorage()
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const toast = useToastStore()

  const [newPath, setNewPath] = useState('')
  const [newToken, setNewToken] = useState('')
  const [matchOhash, setMatchOhash] = useState(false)
  const [videoExt, setVideoExt] = useState<string[]>([])
  const [webhooks, setWebhooks] = useState<WebhooksConfig | null>(null)

  useEffect(() => {
    if (storage) {
      setMatchOhash(storage.match_ohash)
      setVideoExt(storage.video_ext ?? [])
      setWebhooks(storage.webhooks)
    }
  }, [storage])

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['optionsStorage'] })

  const saveOptions = useMutation({
    mutationFn: () => api.put('/options/storage', { match_ohash: matchOhash, video_ext: videoExt, webhooks }),
    onSuccess: () => {
      toast.success('Storage options saved')
      invalidate()
    }
  })

  const addVolume = useMutation({
    mutationFn: (body: { path?: string; token?: string; type: string }) => api.post('/options/storage', body),
    onSuccess: () => {
      setNewPath('')
      setNewToken('')
      invalidate()
    }
  })

  const removeVolume = useMutation({
    mutationFn: (id: number) => api.delete(`/options/storage/${id}`),
    onSuccess: invalidate
  })

  const rescan = (id?: number) => api.get(id ? `/task/rescan/${id}` : '/task/rescan').then(() => toast.info('Rescan started'))

  const volumes: Volume[] = storage?.volumes ?? []
  const forbidden = new Set(storage?.forbidden_video_ext ?? [])

  const sanitizeExt = (ext: string): string | null => {
    let e = ext.trim().toLowerCase()
    if (!e) return null
    if (!e.startsWith('.')) e = `.${e}`
    if (/\s/.test(e) || !/^\.[a-z0-9]+$/.test(e) || forbidden.has(e)) return null
    return e
  }

  const setHook = (key: keyof WebhooksConfig, field: string, value: string) => {
    setWebhooks((w) => (w ? { ...w, [key]: { ...w[key], [field]: value } } : w))
  }

  return (
    <div className="space-y-4">
      <SectionCard
        title="Storage paths"
        actions={
          <button
            onClick={async () => {
              await saveOptions.mutateAsync()
              rescan()
            }}
            className={btnCls}
          >
            Rescan all folders
          </button>
        }
      >
        <table className="mb-3 w-full text-sm">
          <thead>
            <tr className="text-left text-xs uppercase text-muted">
              <th className="py-1 pr-2">Path</th>
              <th className="py-1 pr-2">Type</th>
              <th className="py-1 pr-2">Files</th>
              <th className="py-1 pr-2">Unmatched</th>
              <th className="py-1 pr-2">Size</th>
              <th className="py-1 pr-2">Last scan</th>
              <th className="py-1">Actions</th>
            </tr>
          </thead>
          <tbody>
            {volumes.map((v) => (
              <tr key={v.id} className="border-t border-line">
                <td className="max-w-64 truncate py-1.5 pr-2 font-mono text-xs" title={v.path}>
                  {v.path}
                </td>
                <td className="py-1.5 pr-2 text-xs">{v.type}</td>
                <td className="py-1.5 pr-2 text-xs">{v.file_count}</td>
                <td className="py-1.5 pr-2 text-xs">{v.unmatched_count}</td>
                <td className="py-1.5 pr-2 text-xs">{prettyBytes(v.total_size)}</td>
                <td className="py-1.5 pr-2 text-xs">{formatDate(v.last_scan)}</td>
                <td className="py-1.5">
                  <div className="flex gap-1">
                    <button onClick={() => rescan(v.id)} className={`${btnCls} px-2 py-0.5 text-xs`}>
                      Rescan
                    </button>
                    <button
                      onClick={async () => {
                        if (await askConfirm({ title: `Remove ${v.path}?`, message: 'Files stay on disk.', danger: true }))
                          removeVolume.mutate(v.id)
                      }}
                      className={`${btnDangerCls} px-2 py-0.5 text-xs`}
                    >
                      Remove
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="flex flex-wrap gap-2">
          <input
            value={newPath}
            onChange={(e) => setNewPath(e.target.value)}
            placeholder="/path/to/videos"
            className={`${inputCls} w-72`}
          />
          <button disabled={!newPath.trim()} onClick={() => addVolume.mutate({ path: newPath, type: 'local' })} className={btnCls}>
            Add local folder
          </button>
          <input
            type="password"
            value={newToken}
            onChange={(e) => setNewToken(e.target.value)}
            placeholder="put.io token"
            className={`${inputCls} w-48`}
          />
          <button disabled={!newToken.trim()} onClick={() => addVolume.mutate({ token: newToken, type: 'putio' })} className={btnCls}>
            Add put.io
          </button>
        </div>
      </SectionCard>

      {webhooks && (
        <SectionCard title="Webhooks">
          {(['trigger_external_import', 'refresh_external_import'] as const).map((key) => (
            <div key={key} className="mb-3 rounded-lg border border-line p-3">
              <div className="mb-2 text-sm font-semibold">{key}</div>
              <div className="grid grid-cols-1 gap-2 md:grid-cols-[100px_1fr]">
                <select
                  value={webhooks[key].method}
                  onChange={(e) => setHook(key, 'method', e.target.value)}
                  className={inputCls}
                >
                  {['GET', 'POST', 'PUT'].map((m) => (
                    <option key={m}>{m}</option>
                  ))}
                </select>
                <input
                  value={webhooks[key].url}
                  onChange={(e) => setHook(key, 'url', e.target.value)}
                  placeholder="https://…"
                  className={`${inputCls} font-mono text-xs`}
                />
              </div>
              <textarea
                value={webhooks[key].headers}
                onChange={(e) => setHook(key, 'headers', e.target.value)}
                placeholder="Header: value (one per line)"
                rows={2}
                className={`${inputCls} mt-2 font-mono text-xs`}
              />
            </div>
          ))}
          <p className="mb-2 text-xs text-muted">
            Configured webhooks can be triggered from the navbar “Actions” menu.
          </p>
          <SaveButton onClick={() => saveOptions.mutate()} pending={saveOptions.isPending} label="Save webhooks" />
        </SectionCard>
      )}

      <SectionCard title="Options">
        <Toggle checked={matchOhash} onChange={setMatchOhash} label="Match StashDB hashes" />
        <div className="mt-3">
          <Field label="Video file extensions" hint="Extensions like .funscript/.hsp/.srt are always excluded.">
            <ListEditor items={videoExt} onChange={setVideoExt} addLabel="Add extension" />
          </Field>
          <div className="mt-2 flex gap-2">
            <button onClick={() => setVideoExt(storage?.default_video_ext ?? [])} className={btnCls}>
              Reset to defaults
            </button>
            <SaveButton onClick={() => saveOptions.mutate()} pending={saveOptions.isPending} />
          </div>
        </div>
      </SectionCard>
    </div>
  )
}

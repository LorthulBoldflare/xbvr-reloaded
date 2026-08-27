import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { PreviewQueueStatus } from '../../../api/types'
import { useOptionsState } from '../../../api/hooks'
import { usePreviewsStore } from '../../../store/previews'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { SectionCard, Field, SaveButton, btnCls } from '../common'

// Preview generation settings + test + queue control.
export function PreviewsSection() {
  const { data: state } = useOptionsState()
  const toast = useToastStore.getState()
  const queue = usePreviewsStore((s) => s.queue)
  const previewFn = usePreviewsStore((s) => s.previewFn)
  const clearPreview = usePreviewsStore((s) => s.clearPreview)

  const [form, setForm] = useState({ startTime: 10, snippetLength: 0.4, snippetAmount: 20, resolution: 400, extraSnippet: false })

  useEffect(() => {
    if (state?.config?.library?.preview) setForm(state.config.library.preview)
  }, [state?.config?.library?.preview])

  const { data: status } = useQuery({
    queryKey: ['previewStatus'],
    queryFn: () => api.get<PreviewQueueStatus>('/task/preview/status'),
    refetchInterval: queue?.running ? 2000 : false
  })
  const q = queue ?? status
  const previewURL = useMemo(
    () => (previewFn ? `/api/dms/preview/${previewFn}?ts=${Date.now()}` : ''),
    [previewFn]
  )

  const save = useMutation({
    mutationFn: () => api.put('/options/previews', { ...form, enabled: state?.config?.library?.preview?.enabled ?? true }),
    onSuccess: () => toast.success('Preview settings saved')
  })

  const test = useMutation({
    mutationFn: (regenerate: boolean) => api.post('/options/previews/test', { ...form, enabled: true, regenerate }),
    onSuccess: () => toast.info('Test preview rendering…')
  })

  const slider = (label: string, key: keyof typeof form, min: number, max: number, step: number, fmt: (v: number) => string) => (
    <Field label={`${label}: ${fmt(form[key] as number)}`}>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={form[key] as number}
        onChange={(e) => setForm({ ...form, [key]: Number(e.target.value) })}
        className="w-full"
      />
    </Field>
  )

  return (
    <div className="space-y-4">
      <SectionCard title="Preview snippets">
        <div className="grid max-w-lg grid-cols-1 gap-3">
          {slider('Start time', 'startTime', 5, 60, 1, (v) => `${v}s`)}
          {slider('Snippet length', 'snippetLength', 0.2, 5, 0.2, (v) => `${v.toFixed(1)}s`)}
          {slider('Number of snippets', 'snippetAmount', 2, 40, 1, (v) => `${v}`)}
          {slider('Preview resolution', 'resolution', 300, 800, 10, (v) => `${v}px`)}
          <Toggle
            checked={form.extraSnippet}
            onChange={(v) => setForm({ ...form, extraSnippet: v })}
            label="Grab extra snippet from the end of video"
          />
        </div>
        <div className="mt-3 flex gap-2">
          <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
          <button onClick={() => test.mutate(false)} disabled={test.isPending} className={btnCls}>
            Test settings
          </button>
          <button onClick={() => test.mutate(true)} disabled={test.isPending} className={btnCls}>
            Regenerate test video
          </button>
        </div>
        {previewFn && (
          <div className="mt-3">
            <video key={previewFn} src={previewURL} controls autoPlay muted loop className="max-w-md rounded-xl bg-black" />
            <button onClick={clearPreview} className="mt-1 text-xs text-muted hover:text-fg">
              Dismiss
            </button>
          </div>
        )}
      </SectionCard>

      <SectionCard title="Generation">
        <div className="flex items-center gap-2">
          <button onClick={() => api.get('/task/preview/generate').then(() => toast.info('Preview generation started'))} className={btnCls}>
            Start
          </button>
          <button onClick={() => api.get('/task/preview/stop').then(() => toast.info('Stopping…'))} className={btnCls}>
            Stop
          </button>
          {q && (
            <span className="text-sm text-muted">
              {q.running ? `Running: ${q.completed}/${q.total} (${q.currentScene})` : q.stopping ? 'Stopping…' : 'Not running'}
            </span>
          )}
        </div>
        {q && q.total > 0 && (
          <div className="mt-2 h-2 max-w-md overflow-hidden rounded-full bg-surface-3">
            <div className="h-full bg-accent transition-all" style={{ width: `${(q.completed / q.total) * 100}%` }} />
          </div>
        )}
      </SectionCard>
    </div>
  )
}

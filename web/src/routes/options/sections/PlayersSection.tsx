import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { SectionCard, Field, SaveButton, inputCls } from '../common'
import { buildDeovrPayload } from './deovrPayload'

// Players (DeoVR/HereSphere): shared settings minus credentials (those live
// in the Authentication section), sort sequences, player URLs, remote,
// HereSphere permissions.
export function PlayersSection() {
  const { data: state } = useOptionsState()
  const queryClient = useQueryClient()
  const toast = useToastStore.getState()

  const [form, setForm] = useState<Record<string, any>>({})

  useEffect(() => {
    if (state?.config) {
      const d = state.config.interfaces.deovr
      const h = state.config.interfaces.heresphere
      const p = state.config.interfaces.players
      setForm({
        enabled: d.enabled,
        remote_enabled: d.remote_enabled,
        track_watch_time: d.track_watch_time,
        render_heatmaps: d.render_heatmaps,
        allow_file_deletes: h.allow_file_deletes,
        allow_rating_updates: h.allow_rating_updates,
        allow_favorite_updates: h.allow_favorite_updates,
        allow_hsp_data: h.allow_hsp_data,
        allow_tag_updates: h.allow_tag_updates,
        allow_cuepoint_updates: h.allow_cuepoint_updates,
        allow_watchlist_updates: h.allow_watchlist_updates,
        multitrack_cuepoints: h.multitrack_cuepoints,
        multitrack_cast_cuepoints: h.multitrack_cast_cuepoints,
        retain_non_hsp_cuepoints: h.retain_non_hsp_cuepoints,
        video_sort_seq: p.video_sort_seq,
        script_sort_seq: p.script_sort_seq,
        subtitle_sort_seq: p.subtitle_sort_seq
      })
    }
  }, [state?.config?.interfaces])

  const save = useMutation({
    mutationFn: () => api.put('/options/interface/deovr', buildDeovrPayload(state!.config, form)),
    onSuccess: () => {
      toast.success('Player settings saved')
      queryClient.invalidateQueries({ queryKey: ['optionsState'] })
    }
  })

  const set = (k: string, v: unknown) => setForm((f) => ({ ...f, [k]: v }))
  const t = (key: string, label: string) => (
    <Toggle checked={!!form[key]} onChange={(v) => set(key, v)} label={label} />
  )

  const boundIps = state?.currentState?.server?.bound_ip ?? []

  return (
    <div className="space-y-4">
      <SectionCard title="Shared player settings" actions={<SaveButton onClick={() => save.mutate()} pending={save.isPending} />}>
        <div className="grid max-w-2xl grid-cols-1 gap-2 md:grid-cols-2">
          {t('enabled', 'Enable DeoVR/HereSphere endpoints')}
          {t('remote_enabled', 'Enable DeoVR remote (now playing)')}
          {t('track_watch_time', 'Track watch history')}
          {t('render_heatmaps', 'Render heatmaps in players')}
        </div>
        <p className="mt-2 text-xs text-muted">
          Player credentials live in the <a href="/web/options/authentication" className="text-accent-strong underline">Authentication</a> section.
        </p>
      </SectionCard>

      <SectionCard title="File sort sequences">
        <div className="grid max-w-3xl grid-cols-1 gap-3">
          <Field label="Video files" hint="Comma-separated matchers, first match wins">
            <input value={form.video_sort_seq ?? ''} onChange={(e) => set('video_sort_seq', e.target.value)} className={`${inputCls} font-mono text-xs`} />
          </Field>
          <Field label="Script files">
            <input value={form.script_sort_seq ?? ''} onChange={(e) => set('script_sort_seq', e.target.value)} className={`${inputCls} font-mono text-xs`} />
          </Field>
          <Field label="Subtitle files">
            <input value={form.subtitle_sort_seq ?? ''} onChange={(e) => set('subtitle_sort_seq', e.target.value)} className={`${inputCls} font-mono text-xs`} />
          </Field>
        </div>
        <div className="mt-3">
          <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
        </div>
      </SectionCard>

      <SectionCard title="Player URLs">
        <div className="space-y-1 text-sm">
          {boundIps.map((ip) => (
            <div key={ip} className="font-mono text-xs">
              http://{ip}:{state?.config?.server?.port}/deovr · http://{ip}:{state?.config?.server?.port}/heresphere
            </div>
          ))}
        </div>
      </SectionCard>

      <SectionCard title="HereSphere permissions">
        <div className="grid max-w-2xl grid-cols-1 gap-2 md:grid-cols-2">
          {t('allow_file_deletes', 'Allow file deletes')}
          {t('allow_rating_updates', 'Allow rating updates')}
          {t('allow_favorite_updates', 'Allow favorite updates')}
          {t('allow_hsp_data', 'Allow access to HSP data')}
          {t('allow_tag_updates', 'Allow tag updates')}
          {t('allow_cuepoint_updates', 'Allow cuepoint updates')}
          {t('allow_watchlist_updates', 'Allow watchlist updates')}
          {t('multitrack_cuepoints', 'Multi-track cuepoints')}
          {t('multitrack_cast_cuepoints', 'Multi-track cast cuepoints')}
          {t('retain_non_hsp_cuepoints', 'Keep non-HSP cuepoints')}
        </div>
        <div className="mt-3">
          <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
        </div>
      </SectionCard>
    </div>
  )
}

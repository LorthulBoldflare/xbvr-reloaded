import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { CronSchedule } from '../../../api/types'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { SectionCard, SaveButton } from '../common'

const SCHEDULES: { key: string; label: string; disabledNote?: (state: ReturnType<typeof useOptionsState>['data']) => string | null }[] = [
  { key: 'rescrapeSchedule', label: 'Rescrape' },
  { key: 'rescanSchedule', label: 'Rescan' },
  { key: 'previewSchedule', label: 'Preview generation' },
  { key: 'actorRescrapeSchedule', label: 'Actor rescrape' },
  {
    key: 'stashdbRescrapeSchedule',
    label: 'StashDB rescrape',
    disabledNote: (s) => (s?.config?.advanced?.stashApiKey ? null : 'Requires a StashDB API key (Advanced)')
  },
  { key: 'linkScenesSchedule', label: 'Link scenes' }
]

// Six cron schedules with an identical control set. Note: schedule changes
// require a restart to take effect (same as the old UI).
export function SchedulesSection() {
  const { data: state } = useOptionsState()
  const toast = useToastStore()
  const [form, setForm] = useState<Record<string, CronSchedule>>({})

  useEffect(() => {
    if (state?.config?.cron) setForm(state.config.cron)
  }, [state])

  const save = useMutation({
    mutationFn: () => {
      // Flatten to the RequestSaveOptionsTaskSchedule shape:
      // {task}{Enabled|HourInterval|UseRange|MinuteStart|HourStart|HourEnd|StartDelay}
      const body: Record<string, unknown> = {}
      for (const s of SCHEDULES) {
        const prefix = s.key.replace('Schedule', '')
        const c = form[s.key]
        if (!c) continue
        body[`${prefix}Enabled`] = c.enabled
        body[`${prefix}HourInterval`] = c.hourInterval
        body[`${prefix}UseRange`] = c.useRange
        body[`${prefix}MinuteStart`] = c.minuteStart
        body[`${prefix}HourStart`] = c.hourStart
        body[`${prefix}HourEnd`] = c.hourEnd >= 24 ? c.hourEnd - 24 : c.hourEnd
        body[`${prefix}StartDelay`] = c.runAtStartDelay
      }
      return api.post('/options/task-schedule', body)
    },
    onSuccess: () => toast.success('Schedules saved — a restart is required for changes to take effect')
  })

  const set = (key: string, p: Partial<CronSchedule>) => setForm((f) => ({ ...f, [key]: { ...f[key], ...p } }))

  return (
    <SectionCard
      title="Task schedules"
      actions={<SaveButton onClick={() => save.mutate()} pending={save.isPending} />}
    >
      <div className="grid grid-cols-1 gap-3 xl:grid-cols-2">
        {SCHEDULES.map((s) => {
          const c = form[s.key]
          if (!c) return null
          const disabledNote = s.disabledNote?.(state)
          return (
            <div key={s.key} className={`rounded-lg border border-line p-3 ${disabledNote ? 'opacity-50' : ''}`}>
              <div className="mb-2 flex items-center justify-between">
                <span className="font-medium">{s.label}</span>
                <Toggle checked={c.enabled && !disabledNote} onChange={(v) => set(s.key, { enabled: v })} disabled={!!disabledNote} />
              </div>
              {disabledNote && <div className="mb-2 text-xs text-muted">{disabledNote}</div>}
              <div className="space-y-2 text-xs">
                <label className="block">
                  Run every {c.hourInterval}h
                  <input
                    type="range" min={1} max={23} value={Math.max(1, c.hourInterval)}
                    onChange={(e) => set(s.key, { hourInterval: Number(e.target.value) })}
                    className="w-full"
                  />
                </label>
                <Toggle checked={c.useRange} onChange={(v) => set(s.key, { useRange: v })} label="Limit time of day" />
                {c.useRange && (
                  <div className="flex items-center gap-2">
                    <span className="w-16 text-muted">{c.hourStart}:00–{c.hourEnd}:00</span>
                    <input
                      type="range" min={0} max={23} value={c.hourStart}
                      onChange={(e) => set(s.key, { hourStart: Math.min(Number(e.target.value), c.hourEnd) })}
                      className="w-full"
                    />
                    <input
                      type="range" min={0} max={48} value={c.hourEnd >= c.hourStart ? c.hourEnd : c.hourEnd + 24}
                      onChange={(e) => {
                        let v = Number(e.target.value)
                        if (v - c.hourStart > 24) v = c.hourStart + 24
                        set(s.key, { hourEnd: v })
                      }}
                      className="w-full"
                    />
                  </div>
                )}
                <label className="block">
                  Minute past the hour: {c.minuteStart}
                  <input
                    type="range" min={0} max={60} value={c.minuteStart}
                    onChange={(e) => set(s.key, { minuteStart: Number(e.target.value) })}
                    className="w-full"
                  />
                </label>
                <label className="block">
                  Startup delay: {c.runAtStartDelay} min
                  <input
                    type="range" min={0} max={60} value={c.runAtStartDelay}
                    onChange={(e) => set(s.key, { runAtStartDelay: Number(e.target.value) })}
                    className="w-full"
                  />
                </label>
              </div>
            </div>
          )
        })}
      </div>
    </SectionCard>
  )
}

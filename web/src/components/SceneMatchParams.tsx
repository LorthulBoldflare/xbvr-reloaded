import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import { useToastStore } from '../store/toasts'
import { useUIStore } from '../store/ui'
import { Modal } from './Modal'
import { Field, SaveButton, btnCls, inputCls } from '../routes/options/common'

// Match-parameter editor for sub-sites (old UI: SceneMatchParams overlay).
export function SceneMatchParams({ siteId, onClose }: { siteId: string | null; onClose: () => void }) {
  const toast = useToastStore()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const [params, setParams] = useState<Record<string, any> | null>(null)
  const [dateSet, setDateSet] = useState(false)

  const { data } = useQuery({
    queryKey: ['matchParams', siteId],
    queryFn: () => api.get<Record<string, any>>(`/options/site/match_params/${siteId}`),
    enabled: siteId !== null
  })

  useEffect(() => {
    if (data) {
      setParams({ ...data, ignore_released_before: (data.ignore_released_before ?? '').slice(0, 10) })
      setDateSet(false)
    }
  }, [data])

  const save = useMutation({
    mutationFn: async () => {
      await api.post('/options/site/save_match_params', { site: siteId, match_params: params })
      if (dateSet && params?.ignore_released_before) {
        const clear = await askConfirm({
          title: 'Clear existing links?',
          message: `Delete existing links for this site created before ${params.ignore_released_before}?`
        })
        if (clear) {
          await api.delete('/extref/delete_extref_source_links/keep_manual', {
            external_source: `alternate scene ${siteId}`,
            delete_date: params.ignore_released_before
          })
        }
      }
    },
    onSuccess: () => {
      toast.success('Match parameters saved')
      onClose()
    }
  })

  const set = (k: string, v: unknown) => setParams((p) => ({ ...p!, [k]: v }))

  const matchTypeSelect = (key: string, allowMust: boolean) => (
    <select value={params?.[key] ?? 'should'} onChange={(e) => set(key, e.target.value)} className={inputCls}>
      {allowMust && <option value="must">must</option>}
      <option value="should">should</option>
      <option value="do not">do not</option>
    </select>
  )

  const num = (key: string) => (
    <input
      type="number"
      value={params?.[key] ?? 0}
      onChange={(e) => set(key, Number(e.target.value))}
      className={inputCls}
    />
  )

  const boost = (key: string) => (
    <input
      type="number"
      step={0.25}
      value={params?.[key] ?? 1}
      onChange={(e) => set(key, Number(e.target.value))}
      className={inputCls}
    />
  )

  return (
    <Modal open={siteId !== null} onClose={onClose} width="max-w-2xl" title={`Matching parameters: ${siteId}`}>
      {!params ? (
        <div className="py-8 text-center text-muted">Loading…</div>
      ) : (
        <div className="max-h-[60vh] space-y-4 overflow-y-auto pr-1">
          <div className="grid grid-cols-2 gap-3">
            <Field label="Delay linking (days)">{num('delay_linking')}</Field>
            <Field label="Keep re-linking (days)">{num('reprocess_links')}</Field>
            <Field label="Ignore scenes released prior to">
              <input
                type="date"
                value={params.ignore_released_before ?? ''}
                onChange={(e) => {
                  set('ignore_released_before', e.target.value)
                  setDateSet(true)
                }}
                className={inputCls}
              />
            </Field>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Release date match">{matchTypeSelect('released_match_type', true)}</Field>
            <Field label="Boost release date">{boost('boost_released')}</Field>
            <Field label="Days prior">{num('released_prior')}</Field>
            <Field label="Days after">{num('released_after')}</Field>
            <Field label="Title exact boost">{boost('boost_title')}</Field>
            <Field label="Title any-word boost">{boost('boost_title_any_words')}</Field>
            <Field label="Duration match">{matchTypeSelect('duration_match_type', true)}</Field>
            <Field label="Min duration (min)">{num('duration_min')}</Field>
            <Field label="Duration lower range (s)">{num('duration_range_less')}</Field>
            <Field label="Duration upper range (s)">{num('duration_range_more')}</Field>
            <Field label="Cast match">{matchTypeSelect('cast_match_type', false)}</Field>
            <Field label="Boost cast">{boost('boost_cast')}</Field>
            <Field label="Description match">{matchTypeSelect('desc_match_type', false)}</Field>
            <Field label="Boost description">{boost('boost_description')}</Field>
          </div>
        </div>
      )}
      <div className="mt-4 flex justify-end gap-2">
        <button onClick={onClose} className={btnCls}>
          Cancel
        </button>
        <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
      </div>
    </Modal>
  )
}

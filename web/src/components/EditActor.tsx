import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Actor, CountryDetails, ExternalReferenceLink } from '../api/types'
import { useUIStore } from '../store/ui'
import { useToastStore } from '../store/toasts'
import { useOptionsState } from '../api/hooks'
import { Modal } from './Modal'
import { ListEditor } from './ListEditor'

// Actor editor modal (parity with the old EditActor): information + list
// tabs + scraper URLs; delete only for scene-less non-aka actors.
export function EditActor() {
  const actorId = useUIStore((s) => s.editActorId)
  const hide = useUIStore((s) => s.hideEditActor)
  const askConfirm = useUIStore((s) => s.askConfirm)
  const queryClient = useQueryClient()
  const toast = useToastStore()
  const { data: state } = useOptionsState()
  const imperial = state?.config?.advanced?.useImperialEntry ?? false

  const { data: actor } = useQuery({
    queryKey: ['actor', actorId],
    queryFn: () => api.get<Actor>(`/actor/${actorId}`),
    enabled: actorId !== null
  })
  const { data: countries } = useQuery({
    queryKey: ['countrylist'],
    queryFn: () => api.get<CountryDetails[]>('/actor/countrylist'),
    staleTime: Infinity,
    enabled: actorId !== null
  })
  const { data: extrefs } = useQuery({
    queryKey: ['actorExtrefs', actorId],
    queryFn: () => api.get<ExternalReferenceLink[]>(`/actor/extrefs/${actorId}`),
    enabled: actorId !== null
  })

  const [draft, setDraft] = useState<Record<string, any> | null>(null)
  const [urls, setUrls] = useState<string[]>([])
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (actor) {
      const parse = (s: string) => {
        try {
          return JSON.parse(s || '[]')
        } catch {
          return []
        }
      }
      setDraft({
        name: actor.name,
        birth_date: actor.birth_date?.startsWith('0001-01-01') ? '' : actor.birth_date?.slice(0, 10),
        nationality: actor.nationality,
        ethnicity: actor.ethnicity,
        eye_color: actor.eye_color,
        hair_color: actor.hair_color,
        height: imperial ? Math.round(actor.height / 2.54) : actor.height, // inches when imperial
        weight: imperial ? Math.round(actor.weight * 2.20462) : actor.weight, // lbs when imperial
        cup_size: actor.cup_size,
        band_size: actor.band_size,
        waist_size: actor.waist_size,
        hip_size: actor.hip_size,
        breast_type: actor.breast_type,
        start_year: actor.start_year,
        end_year: actor.end_year,
        biography: actor.biography,
        aliases: parse(actor.aliases),
        tattoos: parse(actor.tattoos),
        piercings: parse(actor.piercings),
        image_arr: parse(actor.image_arr)
      })
      setUrls((extrefs ?? []).map((r) => r.url))
      setDirty(false)
    }
  }, [actor, extrefs, imperial])

  const close = async () => {
    if (dirty && !(await askConfirm({ title: 'Discard unsaved changes?' }))) return
    setDirty(false)
    hide()
  }

  const save = useMutation({
    mutationFn: async () => {
      const d = { ...draft }
      // imperial → metric
      if (imperial) {
        d.height = Math.round(Number(d.height) * 2.54)
        d.weight = Math.round(Number(d.weight) / 2.20462)
      }
      d.aliases = JSON.stringify(d.aliases.filter((x: string) => x.trim() !== ''))
      d.tattoos = JSON.stringify(d.tattoos.filter((x: string) => x.trim() !== ''))
      d.piercings = JSON.stringify(d.piercings.filter((x: string) => x.trim() !== ''))
      d.image_arr = JSON.stringify(d.image_arr.filter((x: string) => x.trim() !== ''))
      await api.post(`/actor/edit/${actorId}`, d)
      await api.post(`/actor/edit_extrefs/${actorId}`, urls.filter((u) => u.trim() !== ''))
    },
    onSuccess: () => {
      toast.success('Actor saved')
      queryClient.invalidateQueries({ queryKey: ['actor', actorId] })
      queryClient.invalidateQueries({ queryKey: ['actorList'] })
      setDirty(false)
      hide()
    }
  })

  const del = useMutation({
    mutationFn: () => api.delete(`/actor/delete/${actorId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actorList'] })
      useUIStore.getState().hideActorDetails()
      hide()
    }
  })

  if (actorId === null) return null

  const set = (k: string, v: unknown) => {
    setDraft((d) => ({ ...d!, [k]: v }))
    setDirty(true)
  }

  const field = (label: string, key: string, type: 'text' | 'number' | 'date' = 'text') => (
    <label className="block">
      <span className="mb-1 block text-xs font-semibold uppercase text-muted">{label}</span>
      <input
        type={type}
        value={draft?.[key] ?? ''}
        onChange={(e) => set(key, type === 'number' ? Number(e.target.value) : e.target.value)}
        className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
      />
    </label>
  )

  return (
    <Modal open onClose={close} width="max-w-3xl" title={`Edit actor: ${actor?.name ?? ''}`}>
      {!draft ? (
        <div className="py-8 text-center text-muted">Loading…</div>
      ) : (
        <div className="max-h-[65vh] space-y-4 overflow-y-auto pr-1">
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
            {field('Name', 'name')}
            <label className="block">
              <span className="mb-1 block text-xs font-semibold uppercase text-muted">Nationality</span>
              <select
                value={draft.nationality}
                onChange={(e) => set('nationality', e.target.value)}
                className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
              >
                <option value="">—</option>
                {(countries ?? []).map((c) => (
                  <option key={c.code} value={c.name}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            {field('Ethnicity', 'ethnicity')}
            {field('Birth date', 'birth_date', 'date')}
            {field('Eye color', 'eye_color')}
            {field('Hair color', 'hair_color')}
            {field(imperial ? 'Height (in)' : 'Height (cm)', 'height', 'number')}
            {field(imperial ? 'Weight (lbs)' : 'Weight (kg)', 'weight', 'number')}
            {field('Cup size', 'cup_size')}
            {field('Band size', 'band_size', 'number')}
            {field('Waist', 'waist_size', 'number')}
            {field('Hip', 'hip_size', 'number')}
            {field('Breast type', 'breast_type')}
            {field('Active from', 'start_year', 'number')}
            {field('Active to', 'end_year', 'number')}
          </div>

          <label className="block">
            <span className="mb-1 block text-xs font-semibold uppercase text-muted">Biography</span>
            <textarea
              value={draft.biography}
              onChange={(e) => set('biography', e.target.value)}
              rows={4}
              className="w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
            />
          </label>

          <div>
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Aliases</div>
            <ListEditor items={draft.aliases} onChange={(v) => set('aliases', v)} />
          </div>
          <div>
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Tattoos</div>
            <ListEditor items={draft.tattoos} onChange={(v) => set('tattoos', v)} />
          </div>
          <div>
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Piercings</div>
            <ListEditor items={draft.piercings} onChange={(v) => set('piercings', v)} />
          </div>
          <div>
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Image URLs</div>
            <ListEditor items={draft.image_arr} onChange={(v) => set('image_arr', v)} showLinks />
          </div>
          <div>
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Scraper URLs</div>
            <ListEditor items={urls} onChange={(v) => { setUrls(v); setDirty(true) }} showLinks />
          </div>
        </div>
      )}

      <div className="mt-4 flex justify-between">
        <div>
          {actor && actor.count === 0 && !actor.name.startsWith('aka:') && (
            <button
              onClick={async () => {
                if (await askConfirm({ title: `Delete actor ${actor.name}?`, danger: true, confirmLabel: 'Delete' }))
                  del.mutate()
              }}
              className="rounded-lg border border-danger/40 px-3 py-1.5 text-sm text-danger hover:bg-danger/10"
            >
              Delete actor
            </button>
          )}
        </div>
        <div className="flex gap-2">
          <button onClick={close} className="rounded-lg border border-line px-3 py-1.5 text-sm">
            Cancel
          </button>
          <button
            disabled={!dirty || save.isPending}
            onClick={() => save.mutate()}
            className="rounded-lg bg-accent px-4 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
          >
            Save details
          </button>
        </div>
      </div>
    </Modal>
  )
}

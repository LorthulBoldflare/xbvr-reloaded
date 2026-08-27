import { useQuery } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { ResponseGetActorFilters } from '../../api/types'
import { normalizeActorFilters, useActorFilterStore } from '../../store/actorFilters'
import { useToastStore } from '../../store/toasts'
import { useUIStore } from '../../store/ui'
import { useQueryClient } from '@tanstack/react-query'
import { akaApi } from '../../api/groups'
import { TagFilterInput } from '../../components/TagFilterInput'
import { SavedSearchPicker } from '../../components/SavedSearchPicker'

export function useActorFilterOptions() {
  return useQuery({
    queryKey: ['actorFilters'],
    queryFn: () => api.get<ResponseGetActorFilters>('/actor/filters'),
    staleTime: 5 * 60_000
  })
}

function RangeFilter({
  label,
  min,
  max,
  lo,
  hi,
  step = 1,
  onChange
}: {
  label: string
  min: number
  max: number
  lo: number
  hi: number
  step?: number
  onChange: (lo: number, hi: number) => void
}) {
  return (
    <div>
      <div className="mb-0.5 flex justify-between text-xs">
        <span className="font-semibold uppercase tracking-wide text-muted">{label}</span>
        <span className="text-muted">
          {lo}–{hi}
        </span>
      </div>
      <div className="flex items-center gap-2">
        <input
          type="range"
          min={min}
          max={max}
          step={step}
          value={lo}
          onChange={(e) => onChange(Math.min(Number(e.target.value), hi), hi)}
          className="w-full"
        />
        <input
          type="range"
          min={min}
          max={max}
          step={step}
          value={hi}
          onChange={(e) => onChange(lo, Math.max(Number(e.target.value), lo))}
          className="w-full"
        />
      </div>
    </div>
  )
}

// Actors filters popover content.
export function ActorFiltersPopoverContent() {
  const { filters, patch, setFilters } = useActorFilterStore()
  const { data: opts } = useActorFilterOptions()
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const toast = useToastStore()

  const reloadFilters = () => queryClient.invalidateQueries({ queryKey: ['actorFilters'] })
  const warn = (status: string) => status && toast.info(`Warning: ${status}`)

  const strip = (v: string) => (v.startsWith('!') || v.startsWith('&') ? v.slice(1) : v)
  const plainCast = filters.cast.filter((c) => !strip(c).startsWith('aka:'))
  const akaSelected = filters.cast.find((c) => strip(c).startsWith('aka:'))

  return (
    <div className="space-y-3">
      <SavedSearchPicker
        type="actor"
        currentFilters={filters as unknown as Record<string, unknown>}
        onApply={(f) => setFilters(normalizeActorFilters(f as never))}
      />

      <div>
        <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Properties</div>
        <div className="flex gap-1.5">
          {[
            { key: 'watchlist', label: 'Watchlist' },
            { key: 'favourite', label: 'Favourite' }
          ].map((p) => (
            <button
              key={p.key}
              onClick={() =>
                patch({ lists: filters.lists.includes(p.key) ? filters.lists.filter((l) => l !== p.key) : [...filters.lists, p.key] })
              }
              className={`rounded-full border px-2.5 py-1 text-xs font-medium ${
                filters.lists.includes(p.key)
                  ? 'border-accent bg-accent-soft text-accent-strong'
                  : 'border-line text-muted hover:text-fg'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      <TagFilterInput label="Cast" values={filters.cast} options={opts?.cast ?? []} onChange={(v) => patch({ cast: v })} mode="2way" />
      <TagFilterInput label="Site" values={filters.sites} options={opts?.sites ?? []} onChange={(v) => patch({ sites: v })} mode="2way" />
      <TagFilterInput
        label="Attributes"
        values={filters.attributes}
        options={opts?.attributes ?? []}
        onChange={(v) => patch({ attributes: v })}
        mode="3way"
      />

      <RangeFilter label="Age" min={18} max={100} lo={Math.max(18, filters.min_age)} hi={filters.max_age} onChange={(lo, hi) => patch({ min_age: lo === 18 ? 0 : lo, max_age: hi })} />
      <RangeFilter label="Height (cm)" min={120} max={220} lo={filters.min_height} hi={filters.max_height} onChange={(lo, hi) => patch({ min_height: lo, max_height: hi })} />
      <RangeFilter label="Weight (kg)" min={25} max={150} lo={filters.min_weight} hi={filters.max_weight} onChange={(lo, hi) => patch({ min_weight: lo, max_weight: hi })} />
      <RangeFilter label="Scenes" min={0} max={150} lo={filters.min_count} hi={filters.max_count} onChange={(lo, hi) => patch({ min_count: lo, max_count: hi })} />
      <RangeFilter label="Available" min={0} max={150} lo={filters.min_avail} hi={filters.max_avail} onChange={(lo, hi) => patch({ min_avail: lo, max_avail: hi })} />
      <RangeFilter label="Rating" min={0} max={5} step={0.5} lo={filters.min_rating} hi={filters.max_rating} onChange={(lo, hi) => patch({ min_rating: lo, max_rating: hi })} />
      <RangeFilter label="Scene rating" min={0} max={5} step={0.25} lo={filters.min_scene_rating} hi={filters.max_scene_rating} onChange={(lo, hi) => patch({ min_scene_rating: lo, max_scene_rating: hi })} />

      <div>
        <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Actor groups (AKA)</div>
        <div className="flex flex-wrap gap-1.5">
          <button
            disabled={plainCast.length < 2}
            onClick={async () => {
              const data = await akaApi.create(filters.cast)
              patch({ cast: [...filters.cast, data.akas.aka_actor.name] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            New
          </button>
          <button
            disabled={!akaSelected}
            onClick={async () => {
              if (!(await askConfirm({ title: `Delete aka group ${strip(akaSelected!)}?`, danger: true }))) return
              const data = await akaApi.delete(strip(akaSelected!))
              patch({ cast: [] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Delete
          </button>
          <button
            disabled={!akaSelected || plainCast.length === 0}
            onClick={async () => {
              const data = await akaApi.add(filters.cast)
              patch({ cast: [...plainCast, data.akas.aka_actor.name] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Add cast
          </button>
          <button
            disabled={!akaSelected || plainCast.length === 0}
            onClick={async () => {
              const data = await akaApi.remove(filters.cast)
              patch({ cast: [...plainCast, data.akas.aka_actor.name] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Remove cast
          </button>
        </div>
      </div>
    </div>
  )
}

import { create } from 'zustand'
import type { ActorFilters } from '../api/types'

export const DEFAULT_ACTOR_FILTERS: ActorFilters = {
  lists: [],
  cast: [],
  sites: [],
  tags: [],
  attributes: [],
  jumpTo: '',
  min_age: 0,
  max_age: 100,
  min_height: 120,
  max_height: 220,
  min_weight: 25,
  max_weight: 150,
  min_count: 0,
  max_count: 150,
  min_avail: 0,
  max_avail: 150,
  min_rating: 0,
  max_rating: 5,
  min_scene_rating: 0,
  max_scene_rating: 5,
  sort: 'name_asc'
}

export const ACTOR_PAGE_SIZE = 24

export const ACTOR_SORTS: { value: string; label: string }[] = [
  { value: 'name_asc', label: '↑ Name' },
  { value: 'name_desc', label: '↓ Name' },
  { value: 'birthday_desc', label: '↓ Birthday' },
  { value: 'birthday_asc', label: '↑ Birthday' },
  { value: 'rating_desc', label: '↓ Rating' },
  { value: 'rating_asc', label: '↑ Rating' },
  { value: 'scene_rating_desc', label: '↓ Scene rating' },
  { value: 'added_desc', label: '↓ Added date' },
  { value: 'added_asc', label: '↑ Added date' },
  { value: 'modified_desc', label: '↓ Modified date' },
  { value: 'modified_asc', label: '↑ Modified date' },
  { value: 'scene_release_desc', label: '↓ Scene release date' },
  { value: 'scene_added_desc', label: '↓ Scene added date' },
  { value: 'file_added_desc', label: '↓ Scene file added date' },
  { value: 'scene_available_desc', label: '↓ Scene available date' },
  { value: 'scene_count_desc', label: '↓ Scene count' },
  { value: 'random', label: '↯ Random' }
]

interface ActorFilterState {
  filters: ActorFilters
  setFilters: (f: ActorFilters) => void
  patch: (p: Partial<ActorFilters>) => void
  reset: () => void
}

export const useActorFilterStore = create<ActorFilterState>((set) => ({
  filters: DEFAULT_ACTOR_FILTERS,
  setFilters: (filters) => set({ filters }),
  patch: (p) => set((s) => ({ filters: { ...s.filters, ...p } })),
  reset: () => set({ filters: DEFAULT_ACTOR_FILTERS })
}))

// Same hardening as normalizeSceneFilters: legacy saved searches / shared
// URLs may carry nulls or wrong types.
export function normalizeActorFilters(p: Partial<ActorFilters> | null | undefined): ActorFilters {
  const d = DEFAULT_ACTOR_FILTERS
  const src = p ?? {}
  const arr = (v: unknown): string[] =>
    Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string') : []
  const num = (v: unknown, def: number): number =>
    typeof v === 'number' && !isNaN(v) ? v : def
  return {
    lists: arr(src.lists),
    cast: arr(src.cast),
    sites: arr(src.sites),
    tags: arr(src.tags),
    attributes: arr(src.attributes),
    jumpTo: typeof src.jumpTo === 'string' ? src.jumpTo : '',
    min_age: num(src.min_age, d.min_age),
    max_age: num(src.max_age, d.max_age),
    min_height: num(src.min_height, d.min_height),
    max_height: num(src.max_height, d.max_height),
    min_weight: num(src.min_weight, d.min_weight),
    max_weight: num(src.max_weight, d.max_weight),
    min_count: num(src.min_count, d.min_count),
    max_count: num(src.max_count, d.max_count),
    min_avail: num(src.min_avail, d.min_avail),
    max_avail: num(src.max_avail, d.max_avail),
    min_rating: num(src.min_rating, d.min_rating),
    max_rating: num(src.max_rating, d.max_rating),
    min_scene_rating: num(src.min_scene_rating, d.min_scene_rating),
    max_scene_rating: num(src.max_scene_rating, d.max_scene_rating),
    sort: typeof src.sort === 'string' && src.sort ? src.sort : d.sort
  }
}

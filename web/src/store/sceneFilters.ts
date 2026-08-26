import { create } from 'zustand'
import type { SceneFilters } from '../api/types'

export const DEFAULT_SCENE_FILTERS: SceneFilters = {
  dlState: 'available',
  isAvailable: true,
  isAccessible: true,
  isHidden: false,
  isWatched: null,
  lists: [],
  cast: [],
  sites: [],
  tags: [],
  cuepoint: [],
  attributes: [],
  volume: 0,
  releaseMonth: '',
  sort: 'file_added_desc',
  showUnmatched: false
}

// Availability presets — port of applyDlState in ui/src/store/sceneList.js.
// Must round-trip through ?q= and saved searches unchanged.
export const DL_STATES: { value: SceneFilters['dlState']; label: string }[] = [
  { value: 'any', label: 'Any' },
  { value: 'available', label: 'Available right now' },
  { value: 'downloaded', label: 'Downloaded' },
  { value: 'missing', label: 'Not downloaded' },
  { value: 'hidden', label: 'Hidden' }
]

export function applyDlState(f: SceneFilters, dl: SceneFilters['dlState']): SceneFilters {
  switch (dl) {
    case 'any':
      return { ...f, dlState: dl, isAvailable: null, isAccessible: null, isHidden: false }
    case 'available':
      return { ...f, dlState: dl, isAvailable: true, isAccessible: true, isHidden: false }
    case 'downloaded':
      return { ...f, dlState: dl, isAvailable: true, isAccessible: null, isHidden: false }
    case 'missing':
      return { ...f, dlState: dl, isAvailable: false, isAccessible: null, isHidden: false }
    case 'hidden':
      return { ...f, dlState: dl, isAvailable: null, isAccessible: null, isHidden: true }
  }
}

interface SceneFilterState {
  filters: SceneFilters
  setFilters: (f: SceneFilters) => void
  patch: (p: Partial<SceneFilters>) => void
  reset: () => void
}

export const useSceneFilterStore = create<SceneFilterState>((set) => ({
  filters: DEFAULT_SCENE_FILTERS,
  setFilters: (filters) => set({ filters }),
  patch: (p) => set((s) => ({ filters: { ...s.filters, ...p } })),
  reset: () => set({ filters: DEFAULT_SCENE_FILTERS })
}))

// Body sent to POST /api/scene/list — strip client-only keys.
export function sceneListRequestBody(f: SceneFilters, offset: number, limit: number) {
  const { showUnmatched: _ignored, ...rest } = f
  return { ...rest, offset, limit }
}

// Keys persisted in ?q= / saved searches (must stay compatible with the old
// UI, including its client-only keys like cardSize which we tolerate on read).
export function filtersForUrl(f: SceneFilters): Record<string, unknown> {
  return { ...f }
}

import { create } from 'zustand'
import type { PreviewQueueStatus } from '../api/types'

// Preview-generation progress + test-preview results (WAMP topics
// options.previews.queue / options.previews.previewReady).

interface PreviewsState {
  queue: PreviewQueueStatus | null
  previewFn: string | null
  setQueue: (q: PreviewQueueStatus) => void
  setPreviewFn: (fn: string) => void
  clearPreview: () => void
}

export const usePreviewsStore = create<PreviewsState>((set) => ({
  queue: null,
  previewFn: null,
  setQueue: (q) => set({ queue: q }),
  setPreviewFn: (fn) => set({ previewFn: fn }),
  clearPreview: () => set({ previewFn: null })
}))

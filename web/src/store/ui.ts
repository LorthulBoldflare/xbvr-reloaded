import { create } from 'zustand'

// Global overlay/modal state (replaces the old UI's store/overlay.js).

export type ConfirmOptions = {
  title: string
  message?: string
  confirmLabel?: string
  danger?: boolean
  resolve: (ok: boolean) => void
}

interface UIState {
  quickFindOpen: boolean
  quickFindForSelect: boolean // when true, selection is stored instead of navigating
  quickFindSelectedScene: unknown | null
  confirm: ConfirmOptions | null
  actorDetailsId: number | null // ActorDetails modal
  editActorId: number | null
  stashdbSceneSearchId: number | null // numeric scene PK
  stashdbActorSearchId: number | null

  openQuickFind: (forSelect?: boolean) => void
  closeQuickFind: () => void
  setQuickFindSelectedScene: (scene: unknown | null) => void
  askConfirm: (opts: Omit<ConfirmOptions, 'resolve'>) => Promise<boolean>
  closeConfirm: (ok: boolean) => void
  showActorDetails: (id: number) => void
  hideActorDetails: () => void
  showEditActor: (id: number) => void
  hideEditActor: () => void
  showStashdbSceneSearch: (sceneId: number) => void
  hideStashdbSceneSearch: () => void
  showStashdbActorSearch: (actorId: number) => void
  hideStashdbActorSearch: () => void
}

export const useUIStore = create<UIState>((set, get) => ({
  quickFindOpen: false,
  quickFindForSelect: false,
  quickFindSelectedScene: null,
  confirm: null,
  actorDetailsId: null,
  editActorId: null,
  stashdbSceneSearchId: null,
  stashdbActorSearchId: null,

  openQuickFind: (forSelect = false) => set({ quickFindOpen: true, quickFindForSelect: forSelect, quickFindSelectedScene: null }),
  closeQuickFind: () => set({ quickFindOpen: false, quickFindForSelect: false }),
  setQuickFindSelectedScene: (scene) => set({ quickFindSelectedScene: scene, quickFindOpen: false }),
  askConfirm: (opts) =>
    new Promise<boolean>((resolve) => {
      set({ confirm: { ...opts, resolve } })
    }),
  closeConfirm: (ok) => {
    get().confirm?.resolve(ok)
    set({ confirm: null })
  },
  showActorDetails: (id) => set({ actorDetailsId: id }),
  hideActorDetails: () => set({ actorDetailsId: null }),
  showEditActor: (id) => set({ editActorId: id }),
  hideEditActor: () => set({ editActorId: null }),
  showStashdbSceneSearch: (sceneId) => set({ stashdbSceneSearchId: sceneId }),
  hideStashdbSceneSearch: () => set({ stashdbSceneSearchId: null }),
  showStashdbActorSearch: (actorId) => set({ stashdbActorSearchId: actorId }),
  hideStashdbActorSearch: () => set({ stashdbActorSearchId: null })
}))

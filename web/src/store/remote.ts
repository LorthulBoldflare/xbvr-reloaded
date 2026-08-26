import { create } from 'zustand'

// DeoVR remote "now playing" state, fed by the WAMP topic `remote.state`.

export interface RemoteState {
  connected: boolean
  deovrHost: string
  isPlaying: boolean
  currentPosition: number
  currentFileID: number
  currentSceneID: string
}

interface RemoteStore extends RemoteState {
  process: (msg: Partial<RemoteState>) => void
}

export const useRemoteStore = create<RemoteStore>((set) => ({
  connected: false,
  deovrHost: '',
  isPlaying: false,
  currentPosition: 0,
  currentFileID: 0,
  currentSceneID: '',
  process: (msg) => set((s) => ({ ...s, ...msg }))
}))

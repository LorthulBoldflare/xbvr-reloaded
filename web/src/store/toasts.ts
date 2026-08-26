import { create } from 'zustand'

export type ToastKind = 'info' | 'success' | 'error'

export interface Toast {
  id: number
  kind: ToastKind
  message: string
}

let nextId = 1

interface ToastState {
  toasts: Toast[]
  push: (kind: ToastKind, message: string) => void
  info: (message: string) => void
  success: (message: string) => void
  error: (message: string) => void
  dismiss: (id: number) => void
}

export const useToastStore = create<ToastState>((set) => ({
  toasts: [],
  push: (kind, message) => {
    const id = nextId++
    set((s) => ({ toasts: [...s.toasts, { id, kind, message }] }))
    setTimeout(() => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })), 5000)
  },
  info: (m) => useToastStore.getState().push('info', m),
  success: (m) => useToastStore.getState().push('success', m),
  error: (m) => useToastStore.getState().push('error', m),
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }))
}))

import { create } from 'zustand'

// Fed by the WAMP websocket (see ws/socket.ts). Mirrors the old UI's
// store/messages.js: scrape/rescan locks, last status messages, and the
// currently-running scrapers.

interface MessagesState {
  lockScrape: boolean
  lockRescan: boolean
  lastScrapeMessage: string
  lastRescanMessage: string
  runningScrapers: string[]
  setLock: (name: string, locked: boolean) => void
  setLastScrapeMessage: (msg: string) => void
  setLastRescanMessage: (msg: string) => void
  addRunningScraper: (id: string) => void
  removeRunningScraper: (id: string) => void
}

export const useMessagesStore = create<MessagesState>((set) => ({
  lockScrape: false,
  lockRescan: false,
  lastScrapeMessage: '',
  lastRescanMessage: '',
  runningScrapers: [],
  setLock: (name, locked) => {
    if (name === 'scrape') set({ lockScrape: locked })
    if (name === 'rescan') set({ lockRescan: locked })
  },
  setLastScrapeMessage: (msg) => set({ lastScrapeMessage: msg }),
  setLastRescanMessage: (msg) => set({ lastRescanMessage: msg }),
  addRunningScraper: (id) =>
    set((s) => (s.runningScrapers.includes(id) ? s : { runningScrapers: [...s.runningScrapers, id] })),
  removeRunningScraper: (id) => set((s) => ({ runningScrapers: s.runningScrapers.filter((x) => x !== id) }))
}))

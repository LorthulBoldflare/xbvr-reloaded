import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { GetStateResponse } from '../api/types'

// Blocking overlay while a DB migration runs. Websocket events keep migration
// progress live; a slow poll while running guarantees convergence if a push
// is missed (pub-sub has no replay, e.g. across a WS reconnect).
export function MigrationOverlay() {
  const { data } = useQuery({
    queryKey: ['optionsState'],
    queryFn: ({ signal }) => api.get<GetStateResponse>('/options/state', { signal, toastOnError: false }),
    staleTime: 60_000,
    refetchInterval: (query) => (query.state.data?.currentState?.migration?.is_running ? 10_000 : false)
  })

  const migration = data?.currentState?.migration
  if (!migration?.is_running) return null

  const pct = migration.total > 0 ? Math.round((migration.progress / migration.total) * 100) : 0

  return (
    <div className="fixed inset-0 z-[90] flex items-center justify-center bg-black/60">
      <div className="w-full max-w-md rounded-xl border border-line bg-surface p-6 shadow-2xl">
        <h2 className="mb-1 text-lg font-semibold">Database Maintenance in Progress</h2>
        <p className="mb-4 text-sm text-muted">{migration.message || migration.current || 'Working…'}</p>
        <div className="h-2 overflow-hidden rounded-full bg-surface-3">
          <div className="h-full bg-accent transition-all" style={{ width: `${pct}%` }} />
        </div>
        <p className="mt-2 text-xs text-muted">
          {migration.progress} / {migration.total}
        </p>
      </div>
    </div>
  )
}

import { useMutation, useQuery } from '@tanstack/react-query'
import { api } from '../../../api/client'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { useUIStore } from '../../../store/ui'
import { prettyBytes } from '../../../lib/format'
import { SectionCard, btnCls } from '../common'

// Cache sizes + resets, search index, scene refresh.
export function CacheSection() {
  const { data: state, refetch } = useOptionsState()
  const toast = useToastStore.getState()
  const askConfirm = useUIStore((s) => s.askConfirm)

  const { data: search, refetch: refetchSearch } = useQuery({
    queryKey: ['searchState'],
    queryFn: () => api.get<{ documentCount: number; inProgress: boolean }>('/options/state/search'),
    refetchInterval: (q) => (q.state.data?.inProgress ? 2000 : false)
  })

  const reset = useMutation({
    mutationFn: (cache: string) => api.delete(`/options/cache/reset/${cache}`),
    onSuccess: () => {
      toast.success('Cache cleared')
      refetch()
      refetchSearch()
    }
  })

  const rows: { label: string; size?: number; cache?: string; extra?: React.ReactNode }[] = [
    { label: 'Images', size: state?.currentState?.cacheSize?.images, cache: 'images' },
    { label: 'Video previews', size: state?.currentState?.cacheSize?.previews, cache: 'previews' },
    {
      label: 'Search index',
      size: state?.currentState?.cacheSize?.searchIndex,
      cache: 'searchIndex',
      extra: (
        <span className="text-xs text-muted">
          {search?.inProgress ? 'Indexing in progress…' : `${search?.documentCount ?? 0} scenes indexed`}
        </span>
      )
    }
  ]

  return (
    <SectionCard title="Cache">
      <div className="space-y-2">
        {rows.map((r) => (
          <div key={r.label} className="flex items-center gap-3 rounded-lg border border-line px-3 py-2">
            <span className="w-32 font-medium">{r.label}</span>
            <span className="text-sm text-muted">{r.size !== undefined ? prettyBytes(r.size) : ''}</span>
            {r.extra}
            <span className="flex-1" />
            {r.cache === 'searchIndex' && (
              <button
                onClick={() => api.get('/task/index').then(() => refetchSearch())}
                className={`${btnCls} px-2 py-0.5 text-xs`}
              >
                Rescan index
              </button>
            )}
            {r.cache && (
              <button
                onClick={async () => {
                  if (await askConfirm({ title: `Reset the ${r.label.toLowerCase()} cache?`, danger: true })) reset.mutate(r.cache!)
                }}
                className={`${btnCls} px-2 py-0.5 text-xs text-danger`}
              >
                Reset
              </button>
            )}
          </div>
        ))}
        <div className="flex items-center gap-3 rounded-lg border border-line px-3 py-2">
          <span className="w-32 font-medium">Scene status</span>
          <span className="flex-1 text-sm text-muted">Recompute available/accessible flags</span>
          <button onClick={() => api.get('/task/scene-refresh').then(() => toast.info('Scene refresh started'))} className={`${btnCls} px-2 py-0.5 text-xs`}>
            Refresh scenes
          </button>
        </div>
        <div className="flex items-center gap-3 rounded-lg border border-line px-3 py-2">
          <span className="w-32 font-medium">Tags</span>
          <span className="flex-1 text-sm text-muted">Remove unused tags</span>
          <button onClick={() => api.get('/task/clean-tags').then(() => toast.info('Tag cleanup started'))} className={`${btnCls} px-2 py-0.5 text-xs`}>
            Clean tags
          </button>
        </div>
      </div>
    </SectionCard>
  )
}

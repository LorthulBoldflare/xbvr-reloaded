import { useEffect, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { api } from '../../../api/client'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { SectionCard, SaveButton } from '../common'

// Funscripts: export links with counts + scrape-for-funscripts switch.
export function FunscriptsSection() {
  const { data: state } = useOptionsState()
  const toast = useToastStore.getState()
  const [scrape, setScrape] = useState(false)

  useEffect(() => {
    setScrape(state?.config?.funscripts?.scrapeFunscripts ?? false)
  }, [state?.config?.funscripts?.scrapeFunscripts])

  const { data: counts, refetch } = useQuery({
    queryKey: ['funscriptCount'],
    queryFn: () => api.get<{ total: number; updated: number }>('/options/funscripts/count')
  })

  const save = useMutation({
    mutationFn: () => api.put('/options/funscripts', { scrapeFunscripts: scrape }),
    onSuccess: () => toast.success('Saved')
  })

  return (
    <SectionCard title="Funscripts">
      <p className="mb-3 text-sm text-muted">
        Export funscripts for use with external tools, or let XBVR scrape for available funscripts during scene
        scrapes.
      </p>
      <div className="mb-4 flex flex-wrap gap-2">
        <a href="/api/task/funscript/export-all" className="rounded-lg border border-line px-3 py-1.5 text-sm hover:bg-surface-2">
          Download all funscripts ({counts?.total ?? 0})
        </a>
        <a
          href="/api/task/funscript/export-new"
          onClick={() => setTimeout(refetch, 3000)}
          className="rounded-lg border border-line px-3 py-1.5 text-sm hover:bg-surface-2"
        >
          Download changes since last export ({counts?.updated ?? 0})
        </a>
      </div>
      <div className="flex items-center gap-3">
        <Toggle checked={scrape} onChange={setScrape} label="Scrape for available funscripts" />
        <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
      </div>
    </SectionCard>
  )
}

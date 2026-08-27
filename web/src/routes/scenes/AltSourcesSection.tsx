import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { AlternateSource, Scene } from '../../api/types'
import { getImageURL, altSourceIconContext } from '../../lib/image'
import { safeHref } from '../../lib/format'
import { useUIStore } from '../../store/ui'
import { useToastStore } from '../../store/toasts'
import { LinkIcon } from '../../components/icons'

// Alternate-source links of a scene: icon + title + external link, plus a
// "Manage link" disclosure with the relink / scrape-and-create / refresh /
// flag-deleted actions (parity with the old Details modal's linked-scene
// actions).
export function AltSourcesSection({ scene }: { scene: Scene }) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const openQuickFindSelect = useUIStore((s) => s.openQuickFind)
  const quickFindSelected = useUIStore((s) => s.quickFindSelectedScene)
  const setQuickFindSelected = useUIStore((s) => s.setQuickFindSelectedScene)
  const toast = useToastStore()
  const [manageOpen, setManageOpen] = useState<string | null>(null)

  const { data } = useQuery({
    queryKey: ['altSources', scene.id],
    queryFn: () => api.get<AlternateSource[]>(`/scene/alternate_source/${scene.id}`)
  })

  const links = (data ?? []).filter(
    (a) => a.external_source.startsWith('alternate scene ') || a.external_source === 'stashdb scene'
  )

  const title = (a: AlternateSource): string => {
    try {
      const ext = JSON.parse(a.external_data)
      return a.external_source.startsWith('alternate scene ') ? (ext.scene?.title ?? 'No title') : (ext.title ?? 'No title')
    } catch {
      return 'No title'
    }
  }

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['altSources', scene.id] })
    queryClient.invalidateQueries({ queryKey: ['scene'] })
  }

  // Relink: user picks a different scene via QuickFind (select mode); the
  // effect below performs the edit_link call after selection.
  const [relinkTarget, setRelinkTarget] = useState<AlternateSource | null>(null)

  const startRelink = (a: AlternateSource) => {
    setRelinkTarget(a)
    openQuickFindSelect(true)
  }

  // Watch for a selection while a relink is pending.
  if (relinkTarget && quickFindSelected) {
    const selected = quickFindSelected as Scene
    const target = relinkTarget
    setRelinkTarget(null)
    setQuickFindSelected(null)
    askConfirm({
      title: 'Relink alternate source',
      message: `Link "${title(target)}" to scene "${selected.title}" instead?`
    }).then(async (ok) => {
      if (!ok) return
      await api.post('/extref/edit_link', {
        external_source: target.external_source,
        external_id: target.external_id,
        internal_table: 'scenes',
        internal_db_id: selected.id,
        internal_name_id: selected.scene_id,
        match_type: 99999
      })
      toast.success('Link updated')
      invalidate()
    })
  }

  const scrapeAndCreate = async (a: AlternateSource) => {
    const ok = await askConfirm({
      title: 'Scrape and create scene',
      message: `Scrape "${title(a)}" as its own scene? You can edit it afterwards.`
    })
    if (ok) navigate(`/options/create?url=${encodeURIComponent(a.url)}`)
  }

  const refresh = async (a: AlternateSource) => {
    const ok = await askConfirm({
      title: 'Refresh external reference',
      message: 'Delete this external reference? It will be re-created on the next scrape with fresh data.'
    })
    if (!ok) return
    await api.delete('/extref/delete_extref', { external_source: a.external_source, external_id: a.external_id })
    toast.success('External reference deleted')
    invalidate()
  }

  const flagDeleted = async (a: AlternateSource) => {
    const ok = await askConfirm({
      title: 'Flag as deleted',
      message: 'Marks this external scene as deleted and unlinks it. This cannot be undone.',
      danger: true
    })
    if (!ok) return
    await api.post('/extref/edit_link', {
      external_source: a.external_source,
      external_id: a.external_id,
      internal_table: 'scenes',
      internal_db_id: 0,
      internal_name_id: 'deleted',
      match_type: -1
    })
    toast.success('Flagged as deleted')
    invalidate()
  }

  if (links.length === 0) return null

  return (
    <div className="space-y-1">
      {links.map((a) => (
        <div key={`${a.external_source}:${a.external_id}`} className="rounded-lg border border-line bg-surface-2 px-3 py-2">
          <div className="flex items-center gap-2">
            <img
              src={getImageURL(a.site_icon, '20x', altSourceIconContext(a))}
              alt=""
              className="h-5 w-5 rounded"
              onError={(e) => ((e.target as HTMLImageElement).style.display = 'none')}
            />
            <span className="min-w-0 flex-1 truncate text-sm">{title(a)}</span>
            <span className="text-xs text-muted">{a.external_source.replace('alternate scene ', '').replace('stashdb scene', 'stashdb')}</span>
            <a href={safeHref(a.url)} target="_blank" rel="noreferrer" className="text-muted hover:text-accent" title="Open external page">
              <LinkIcon />
            </a>
            <button
              onClick={() => setManageOpen(manageOpen === a.external_id ? null : a.external_id)}
              className="rounded border border-line px-2 py-0.5 text-xs text-muted hover:text-fg"
            >
              Manage link
            </button>
          </div>
          {manageOpen === a.external_id && (
            <div className="mt-2 flex flex-wrap gap-1.5 border-t border-line pt-2">
              <button onClick={() => startRelink(a)} className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-3">
                Link to a different scene
              </button>
              <button onClick={() => scrapeAndCreate(a)} className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-3">
                Scrape and create scene
              </button>
              <button onClick={() => refresh(a)} className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-3">
                Refresh external reference
              </button>
              <button
                onClick={() => flagDeleted(a)}
                className="rounded-lg border border-danger/40 px-2 py-1 text-xs text-danger hover:bg-danger/10"
              >
                Flag as deleted
              </button>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

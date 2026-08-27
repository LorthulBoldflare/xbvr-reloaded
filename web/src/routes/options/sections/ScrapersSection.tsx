import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { Scene, Site } from '../../../api/types'
import { formatDate } from '../../../lib/format'
import { useToastStore } from '../../../store/toasts'
import { useUIStore } from '../../../store/ui'
import { useMessagesStore } from '../../../store/messages'
import { useOptionsState } from '../../../api/hooks'
import { getImageURL, iconSlug } from '../../../lib/image'
import { Toggle } from '../../../components/Toggle'
import { Modal } from '../../../components/Modal'
import { Popover } from '../../../components/Popover'
import { SceneMatchParams } from '../../../components/SceneMatchParams'
import { SectionCard, Field, btnCls, inputCls } from '../common'

// Scrapers: site table with enable/limit/subscribed/stash toggles, per-row
// scrape actions, bulk toggles, and the match-params editor for sub-sites.
export function ScrapersSection() {
  const queryClient = useQueryClient()
  const toast = useToastStore()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const runningScrapers = useMessagesStore((s) => s.runningScrapers)
  const { data: state } = useOptionsState()

  const [enabledOnly, setEnabledOnly] = useState(true)
  const [matchParamsSite, setMatchParamsSite] = useState<string | null>(null)
  const [singleScrapeSite, setSingleScrapeSite] = useState<Site | null>(null)

  const { data: sites } = useQuery({
    queryKey: ['sites'],
    queryFn: () => api.get<Site[]>('/options/sites')
  })

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['sites'] })

  const toggleField = useMutation({
    mutationFn: ({ url }: { url: string }) => api.put(url, undefined),
    onSuccess: invalidate
  })

  // NOTE: the state endpoint redacts stashApiKey to "***"; the server keeps
  // the stored key when it receives that sentinel — always round-trip it.
  const saveAutoLimit = useMutation({
    mutationFn: (autoLimitScraping: boolean) =>
      api.put('/options/interface/advanced', {
        ...state?.config?.advanced,
        autoLimitScraping
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['optionsState'] })
      toast.success('Saved')
    }
  })

  const bulkToggle = useMutation({
    mutationFn: ({ field, ids }: { field: string; ids: string[] }) =>
      api.put('/options/sites/toggle_field', { field, ids }),
    onSuccess: invalidate
  })

  const masterSites = useMemo(() => (sites ?? []).filter((s) => s.master_site_id === '' || s.master_site_id === s.id), [sites])
  const visible = useMemo(
    () => (sites ?? []).filter((s) => !enabledOnly || s.is_enabled),
    [sites, enabledOnly]
  )

  const source = (name: string) => {
    const m = name.match(/\(([^)]+)\)\s*$/)
    return m ? m[1] : ''
  }
  const studio = (name: string) => name.replace(/\s*\([^)]+\)\s*$/, '')

  const runScraper = (site: string) => api.get(`/task/scrape?site=${encodeURIComponent(site)}`).then(() => toast.info('Scrape started'))

  return (
    <SectionCard
      title="Scrapers"
      actions={
        <div className="flex items-center gap-3">
          <Toggle
            checked={state?.config?.advanced?.autoLimitScraping ?? false}
            onChange={(v) => saveAutoLimit.mutate(v)}
            label="Auto limit scraping"
          />
          <Toggle checked={enabledOnly} onChange={setEnabledOnly} label="Show enabled only" />
          <button onClick={() => runScraper('_enabled')} className={btnCls}>
            Run enabled scrapers
          </button>
        </div>
      }
    >
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-xs uppercase text-muted">
              <th className="py-1 pr-2">Enabled</th>
              <th className="py-1 pr-2" />
              <th className="py-1 pr-2">Studio</th>
              <th className="py-1 pr-2">Source</th>
              <th className="py-1 pr-2">Last scrape</th>
              <th className="py-1 pr-2">Limit</th>
              <th className="py-1 pr-2">Subscribed</th>
              <th className="py-1 pr-2">Scrape Stash</th>
              <th className="py-1 pr-2">Scenes</th>
              <th className="py-1 pr-2">Main site</th>
              <th className="py-1" />
            </tr>
          </thead>
          <tbody>
            {visible.map((s) => (
              <tr key={s.id} className="border-t border-line">
                <td className="py-1.5 pr-2">
                  <Toggle checked={s.is_enabled} onChange={() => toggleField.mutate({ url: `/options/sites/${s.id}` })} />
                </td>
                <td className="py-1.5 pr-2">
                  {s.avatar_url && <img src={getImageURL(s.avatar_url, '128x', 'icon-' + iconSlug(s.id))} alt="" className="h-6 w-6 rounded" loading="lazy" />}
                </td>
                <td className={`py-1.5 pr-2 font-medium ${s.has_scraper ? '' : 'text-danger'}`} title={s.has_scraper ? '' : 'No scraper'}>
                  {studio(s.name)}
                </td>
                <td className="py-1.5 pr-2 text-xs text-muted">{source(s.name)}</td>
                <td className="py-1.5 pr-2 text-xs">
                  {runningScrapers.includes(s.id) ? (
                    <span className="animate-pulse text-accent-strong">scraping…</span>
                  ) : (
                    <span className="text-muted">{formatDate(s.last_update)}</span>
                  )}
                </td>
                <td className="py-1.5 pr-2">
                  <Toggle checked={s.limit_scraping} onChange={() => toggleField.mutate({ url: `/options/sites/limit_scraping/${s.id}` })} />
                </td>
                <td className="py-1.5 pr-2">
                  {(s.master_site_id === '' || s.master_site_id === s.id) && (
                    <Toggle checked={s.subscribed} onChange={() => toggleField.mutate({ url: `/options/sites/subscribed/${s.id}` })} />
                  )}
                </td>
                <td className="py-1.5 pr-2">
                  <Toggle checked={s.scrape_stash} onChange={() => toggleField.mutate({ url: `/options/sites/scrape_stash/${s.id}` })} />
                </td>
                <td className="py-1.5 pr-2 text-xs">{s.scene_count}</td>
                <td className="py-1.5 pr-2 text-xs text-muted">
                  {s.master_site_id && s.master_site_id !== s.id && (
                    <span className="flex items-center gap-1">
                      {s.master_site_id}
                      <button onClick={() => setMatchParamsSite(s.id)} className="hover:text-fg" title="Matching parameters">
                        ⚙
                      </button>
                    </span>
                  )}
                </td>
                <td className="py-1.5">
                  <Popover button={<span className="text-muted">⋯</span>} align="right" width="w-64" buttonClassName="rounded px-2 py-1 hover:bg-surface-2">
                    {(close) => (
                      <div className="flex flex-col text-sm">
                        <button
                          className="rounded px-2 py-1.5 text-left hover:bg-accent-soft"
                          onClick={() => {
                            runScraper(s.id)
                            close()
                          }}
                        >
                          Run this scraper
                        </button>
                        <button
                          className="rounded px-2 py-1.5 text-left hover:bg-accent-soft"
                          onClick={() => {
                            setSingleScrapeSite(s)
                            close()
                          }}
                        >
                          Scrape single scene
                        </button>
                        <button
                          className="rounded px-2 py-1.5 text-left hover:bg-accent-soft"
                          onClick={async () => {
                            close()
                            if (await askConfirm({ title: `Force update all scenes of ${studio(s.name)}?` })) {
                              await api.post('/options/scraper/force-site-update', { scraper_id: s.id })
                              toast.success('Scenes marked for update')
                            }
                          }}
                        >
                          Force update scenes
                        </button>
                        {s.master_site_id !== '' && s.master_site_id !== s.id && (
                          <>
                            <button
                              className="rounded px-2 py-1.5 text-left hover:bg-accent-soft"
                              onClick={async () => {
                                close()
                                if (await askConfirm({ title: 'Remove scene links (including manual edits)?', danger: true })) {
                                  await api.delete('/extref/delete_extref_source_links/all', { external_source: `alternate scene ${s.id}` })
                                  toast.success('Links removed')
                                }
                              }}
                            >
                              Remove scene links
                            </button>
                            <button
                              className="rounded px-2 py-1.5 text-left hover:bg-accent-soft"
                              onClick={async () => {
                                close()
                                if (await askConfirm({ title: 'Remove scene links (keep manual edits)?' })) {
                                  await api.delete('/extref/delete_extref_source_links/keep_manual', { external_source: `alternate scene ${s.id}` })
                                  toast.success('Links removed')
                                }
                              }}
                            >
                              Remove scene links (keep edits)
                            </button>
                          </>
                        )}
                        <button
                          className="rounded px-2 py-1.5 text-left text-danger hover:bg-danger/10"
                          onClick={async () => {
                            close()
                            if (!(await askConfirm({ title: `Delete all scraped scenes of ${studio(s.name)}?`, danger: true }))) return
                            if (s.master_site_id === '' || s.master_site_id === s.id) {
                              await api.post('/options/scraper/delete-scenes', { scraper_id: s.id })
                            } else {
                              await api.delete('/extref/delete_extref_source', { external_source: `alternate scene ${s.id}` })
                            }
                            toast.success('Scenes deleted')
                            queryClient.invalidateQueries({ queryKey: ['sceneList'] })
                          }}
                        >
                          Delete scraped scenes
                        </button>
                        <button
                          className="rounded px-2 py-1.5 text-left hover:bg-accent-soft"
                          onClick={() => {
                            api.get(`/extref/generic/scrape_by_site/${s.id}`).then(() => toast.info('Actor scrape started'))
                            close()
                          }}
                        >
                          Scrape actor details from site
                        </button>
                      </div>
                    )}
                  </Popover>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-3 flex gap-2 border-t border-line pt-3">
        <button
          onClick={() => bulkToggle.mutate({ field: 'LimitScraping', ids: visible.map((s) => s.id) })}
          className={btnCls}
        >
          Toggle limit scraping (visible)
        </button>
        <button
          onClick={() => bulkToggle.mutate({ field: 'Subscribed', ids: visible.map((s) => s.id) })}
          className={btnCls}
        >
          Toggle subscriptions (visible)
        </button>
      </div>

      <SceneMatchParams siteId={matchParamsSite} onClose={() => setMatchParamsSite(null)} />
      <SingleScrapeModal site={singleScrapeSite} onClose={() => setSingleScrapeSite(null)} />
    </SectionCard>
  )
}

// "Scrape single scene" dialog (collects the scene URL + optional extra info).
function SingleScrapeModal({ site, onClose }: { site: Site | null; onClose: () => void }) {
  const toast = useToastStore()
  const [url, setUrl] = useState('')
  const [extra, setExtra] = useState('')
  const [busy, setBusy] = useState(false)

  useState(() => {
    setUrl('')
    setExtra('')
  })

  const scrape = async () => {
    if (!site || !url.trim()) return
    setBusy(true)
    try {
      // long-running: fire and await the synchronous response (no timeout on server)
      const res = await api.post<{ status: string; scene?: Scene }>('/task/singlescrape', {
        site: site.id,
        sceneurl: url,
        additionalinfo: extra.trim() ? [extra.trim()] : []
      })
      if (res?.scene) {
        toast.success('Scene scraped')
        onClose()
      } else {
        toast.error(res?.status || 'Scrape failed')
      }
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal open={site !== null} onClose={onClose} width="max-w-md" title={`Scrape single scene: ${site?.name ?? ''}`}>
      <Field label="Scene URL">
        <input autoFocus value={url} onChange={(e) => setUrl(e.target.value)} className={`${inputCls} font-mono text-xs`} />
      </Field>
      {site?.id === 'wetvr' && (
        <div className="mt-2">
          <Field label="WetVR scene id">
            <input value={extra} onChange={(e) => setExtra(e.target.value)} className={inputCls} />
          </Field>
        </div>
      )}
      <div className="mt-4 flex justify-end gap-2">
        <button onClick={onClose} className={btnCls}>
          Cancel
        </button>
        <button onClick={scrape} disabled={!url.trim() || busy} className="rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50">
          {busy ? 'Scraping…' : 'Scrape'}
        </button>
      </div>
    </Modal>
  )
}

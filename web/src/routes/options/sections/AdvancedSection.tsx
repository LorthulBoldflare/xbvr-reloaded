import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { CollectorConfig, Site } from '../../../api/types'

// GET /api/options/collector-config-list returns full configs, not just keys.
interface CollectorConfigEntry {
  domain_key: string
  config: CollectorConfig
}
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { useUIStore } from '../../../store/ui'
import { Toggle } from '../../../components/Toggle'
import { ListEditor } from '../../../components/ListEditor'
import { SectionCard, Field, SaveButton, btnCls, inputCls } from '../common'

// Advanced: scene-details switches, actor scraping settings, custom sites,
// alternate-site linking, scraper proxy, per-domain cookies/headers.
export function AdvancedSection() {
  const { data: state } = useOptionsState()
  const queryClient = useQueryClient()
  const toast = useToastStore.getState()
  const askConfirm = useUIStore((s) => s.askConfirm)

  const [form, setForm] = useState<Record<string, any> | null>(null)

  useEffect(() => {
    if (state?.config?.advanced) {
      setForm({
        ...state.config.advanced,
        ignoreReleasedBefore: (state.config.advanced.ignoreReleasedBefore ?? '').slice(0, 10)
      })
    }
  }, [state])

  const { data: sites } = useQuery({ queryKey: ['sites'], queryFn: () => api.get<Site[]>('/options/sites') })
  const { data: collectorList } = useQuery({
    queryKey: ['collectorConfigs'],
    queryFn: () => api.get<CollectorConfigEntry[]>('/options/collector-config-list')
  })

  const save = useMutation({
    mutationFn: () => api.put('/options/interface/advanced', form),
    onSuccess: () => {
      toast.success('Advanced settings saved')
      queryClient.invalidateQueries({ queryKey: ['optionsState'] })
    }
  })

  // collector config editor state
  const [collectorKey, setCollectorKey] = useState('')
  const [collector, setCollector] = useState<CollectorConfig | null>(null)
  const saveCollector = useMutation({
    mutationFn: () => api.post('/options/save-collector-config', collector),
    onSuccess: () => {
      toast.success('Collector config saved')
      queryClient.invalidateQueries({ queryKey: ['collectorConfigs'] })
    }
  })

  // custom site form
  const [newSite, setNewSite] = useState({ scraperUrl: '', scraperName: '', scraperAvatar: '', scraperCompany: '', masterSiteId: '' })

  if (!form) return null
  const set = (k: string, v: unknown) => setForm((f) => ({ ...f!, [k]: v }))

  const subSites = (sites ?? []).filter((s) => s.master_site_id !== '' && s.master_site_id !== s.id)

  const clearLinks = async (siteId: string, keepManual: boolean) => {
    const ok = await askConfirm({
      title: keepManual ? 'Remove links (keep manual edits)?' : 'Remove ALL links (incl. manual edits)?',
      danger: !keepManual
    })
    if (!ok) return
    await api.delete(`/extref/delete_extref_source_links/${keepManual ? 'keep_manual' : 'all'}`, {
      external_source: `alternate scene ${siteId}`
    })
    toast.success('Links removed')
  }

  const loadCollector = (key: string) => {
    setCollectorKey(key)
    if (!key) {
      setCollector(null)
      return
    }
    const existing = (collectorList ?? []).find((c) => c.domain_key === key)
    setCollector(
      existing
        ? { ...existing.config, domain_key: key }
        : { domain_key: key, headers: [], cookies: [], body: '', other: [] }
    )
  }

  return (
    <div className="space-y-4">
      <SectionCard title="Scene details" actions={<SaveButton onClick={() => save.mutate()} pending={save.isPending} />}>
        <div className="grid max-w-3xl grid-cols-1 gap-2 md:grid-cols-2">
          <Toggle checked={form.showInternalSceneId} onChange={(v) => set('showInternalSceneId', v)} label="Show internal scene id" />
          <Toggle checked={form.showHSPApiLink} onChange={(v) => set('showHSPApiLink', v)} label="Show HereSphere API link" />
          <Toggle checked={form.showSceneSearchField} onChange={(v) => set('showSceneSearchField', v)} label="Show scene search fields" />
        </div>
        <p className="mt-2 text-xs text-muted">
          The MCP token moved to the <a href="/web/options/authentication" className="text-accent-strong underline">Authentication</a> section.
        </p>
      </SectionCard>

      <SectionCard title="Actor scraping">
        <div className="max-w-md space-y-3">
          <Field label="StashDB API key" hint="Shown as *** when set — enter a new key to change it">
            <input
              type="password"
              value={form.stashApiKey ?? ''}
              onChange={(e) => set('stashApiKey', e.target.value)}
              className={inputCls}
              autoComplete="off"
            />
          </Field>
          <Toggle checked={form.scrapeActorAfterScene} onChange={(v) => set('scrapeActorAfterScene', v)} label="Scrape actors after scene scrape" />
          <Toggle checked={form.useImperialEntry} onChange={(v) => set('useImperialEntry', v)} label="Enter measurements in imperial units" />
          <div className="flex flex-wrap gap-2">
            <button
              onClick={() => api.get('/extref/stashdb/run_all').then(() => toast.info('StashDB scrape started'))}
              className={btnCls}
            >
              Scrape all actors from StashDB
            </button>
            <button
              onClick={() => api.get('/extref/generic/scrape_all').then(() => toast.info('Actor scrape started'))}
              className={btnCls}
            >
              Scrape all actors from sites
            </button>
          </div>
          <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
        </div>
      </SectionCard>

      <SectionCard title="Alternate sites">
        <div className="mb-3 grid max-w-3xl grid-cols-1 gap-2 md:grid-cols-2">
          <Toggle
            checked={form.linkScenesAfterSceneScraping}
            onChange={(v) => set('linkScenesAfterSceneScraping', v)}
            label="Link scenes after scraping"
          />
          <Toggle
            checked={form.useAltSrcInFileMatching}
            onChange={(v) => set('useAltSrcInFileMatching', v)}
            label="Use alternate sources in file matching"
          />
          <Toggle
            checked={form.useAltSrcInScriptFilters}
            onChange={(v) => set('useAltSrcInScriptFilters', v)}
            label="Use alternate sources in script filters"
          />
        </div>
        <div className="flex max-w-md flex-wrap items-end gap-2">
          <Field label="Ignore scenes released before">
            <input
              type="date"
              value={form.ignoreReleasedBefore ?? ''}
              onChange={(e) => set('ignoreReleasedBefore', e.target.value)}
              className={inputCls}
            />
          </Field>
          <button
            onClick={() => api.get('/task/relink_alt_aource_scenes').then(() => toast.info('Relink started'))}
            className={btnCls}
            title="Re-run linking for alternate-source scenes"
          >
            Relink alternate scenes
          </button>
        </div>
        {subSites.length > 0 && (
          <div className="mt-3">
            <div className="mb-1 text-xs font-semibold uppercase text-muted">Clear links per sub-site</div>
            <div className="flex flex-wrap gap-1.5">
              {subSites.map((s) => (
                <span key={s.id} className="flex items-center gap-1 rounded-lg border border-line px-2 py-1 text-xs">
                  {s.name}
                  <button onClick={() => clearLinks(s.id, true)} className="text-muted hover:text-fg" title="Clear (keep manual)">
                    clear
                  </button>
                  <button onClick={() => clearLinks(s.id, false)} className="text-muted hover:text-danger" title="Clear all">
                    clear!
                  </button>
                </span>
              ))}
            </div>
          </div>
        )}
        <div className="mt-3">
          <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
        </div>
      </SectionCard>

      <SectionCard title="Create custom site">
        <div className="grid max-w-3xl grid-cols-1 gap-3 md:grid-cols-2">
          <Field label="URL">
            <input value={newSite.scraperUrl} onChange={(e) => setNewSite({ ...newSite, scraperUrl: e.target.value })} className={inputCls} />
          </Field>
          <Field label="Name">
            <input value={newSite.scraperName} onChange={(e) => setNewSite({ ...newSite, scraperName: e.target.value })} className={inputCls} />
          </Field>
          <Field label="Avatar URL">
            <input value={newSite.scraperAvatar} onChange={(e) => setNewSite({ ...newSite, scraperAvatar: e.target.value })} className={inputCls} />
          </Field>
          <Field label="Company">
            <input value={newSite.scraperCompany} onChange={(e) => setNewSite({ ...newSite, scraperCompany: e.target.value })} className={inputCls} />
          </Field>
          <Field label="Main site">
            <select
              value={newSite.masterSiteId}
              onChange={(e) => setNewSite({ ...newSite, masterSiteId: e.target.value })}
              className={inputCls}
            >
              <option value="">—</option>
              {(sites ?? [])
                .filter((s) => s.master_site_id === '' || s.master_site_id === s.id)
                .map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
            </select>
          </Field>
        </div>
        <button
          className={`${btnCls} mt-3`}
          disabled={!newSite.scraperUrl.trim() || !newSite.scraperName.trim()}
          onClick={async () => {
            await api.put('/options/custom-sites/create', newSite)
            toast.success('Site created')
            setNewSite({ scraperUrl: '', scraperName: '', scraperAvatar: '', scraperCompany: '', masterSiteId: '' })
            queryClient.invalidateQueries({ queryKey: ['sites'] })
          }}
        >
          Create site
        </button>
      </SectionCard>

      <SectionCard title="Scraper proxy" actions={<SaveButton onClick={() => save.mutate()} pending={save.isPending} />}>
        <div className="max-w-md">
          <input
            value={form.scraperProxy ?? ''}
            onChange={(e) => set('scraperProxy', e.target.value)}
            placeholder="http://proxy:8080"
            className={`${inputCls} font-mono text-xs`}
          />
        </div>
      </SectionCard>

      <SectionCard title="Cookies &amp; headers (per domain)">
        <div className="mb-2 max-w-md">
          <select value={collectorKey} onChange={(e) => loadCollector(e.target.value)} className={inputCls}>
            <option value="">— select domain —</option>
            {(collectorList ?? []).map((c) => (
              <option key={c.domain_key} value={c.domain_key}>
                {c.domain_key}
              </option>
            ))}
          </select>
          <div className="mt-1 flex gap-1">
            <input
              value={collectorKey}
              onChange={(e) => setCollectorKey(e.target.value)}
              placeholder="or type a domain (e.g. stashdb.org)"
              className={`${inputCls} font-mono text-xs`}
            />
            <button onClick={() => loadCollector(collectorKey)} disabled={!collectorKey.trim()} className={btnCls}>
              Load
            </button>
          </div>
        </div>
        {collector && (
          <div className="space-y-3">
            <Field label="Headers">
              <ListEditor
                items={collector.headers.map((h) => `${h.name}: ${h.value}`)}
                onChange={(v) =>
                  setCollector({
                    ...collector,
                    headers: v.filter((x) => x.trim()).map((x) => {
                      const idx = x.indexOf(':')
                      return { name: x.slice(0, idx).trim(), value: x.slice(idx + 1).trim() }
                    })
                  })
                }
                addLabel="Add header"
              />
            </Field>
            <Field label="Cookies (name=value)">
              <ListEditor
                items={collector.cookies.map((c) => `${c.name}=${c.value}`)}
                onChange={(v) =>
                  setCollector({
                    ...collector,
                    cookies: v.filter((x) => x.trim()).map((x) => {
                      const idx = x.indexOf('=')
                      return { name: x.slice(0, idx).trim(), value: x.slice(idx + 1).trim(), domain: collector.domain_key, path: '/', host: '' }
                    })
                  })
                }
                addLabel="Add cookie"
              />
            </Field>
            <div className="flex gap-2">
              <SaveButton onClick={() => saveCollector.mutate()} pending={saveCollector.isPending} />
              <button
                onClick={async () => {
                  if (await askConfirm({ title: `Delete collector config for ${collector.domain_key}?`, danger: true })) {
                    await api.delete('/options/delete-collector-config', { domain_key: collector.domain_key })
                    setCollector(null)
                    setCollectorKey('')
                    queryClient.invalidateQueries({ queryKey: ['collectorConfigs'] })
                  }
                }}
                className={`${btnCls} text-danger`}
              >
                Delete
              </button>
            </div>
          </div>
        )}
      </SectionCard>
    </div>
  )
}

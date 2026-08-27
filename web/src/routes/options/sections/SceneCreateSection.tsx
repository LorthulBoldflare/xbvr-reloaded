import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { Scene, Site } from '../../../api/types'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { SectionCard, Field, btnPrimaryCls, inputCls } from '../common'

const JAVR_SCRAPERS = ['javdatabase', 'r18d', 'javlibrary', 'javland']

const SINGLE_SCENE_SITES: { match: string; site: string; warn?: string }[] = [
  { match: 'sexlikereal.com', site: 'slr-single_scene' },
  { match: 'czechvrnetwork.com', site: 'czechvr-single_scene' },
  { match: 'povr.com', site: 'povr-single_scene' },
  { match: 'vrporn.com', site: 'vrporn-single_scene' },
  { match: 'vrphub.com', site: 'vrphub-single_scene' },
  { match: 'realvr.com', site: 'realvr-single_scene' },
  { match: 'stashdb.org', site: 'stashdb' }
]

// Create / import scene: JAVR, TPDB, custom scene, scrape-by-URL.
export function SceneCreateSection() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const toast = useToastStore.getState()
  const { data: state } = useOptionsState()

  const [javrScraper, setJavrScraper] = useState('javdatabase')
  const [javrId, setJavrId] = useState('')
  const [tpdbUrl, setTpdbUrl] = useState('')
  const [tpdbToken, setTpdbToken] = useState('')
  const [customTitle, setCustomTitle] = useState('')
  const [customId, setCustomId] = useState('')
  const [scrapeUrl, setScrapeUrl] = useState(searchParams.get('url') ?? '')
  const [wetvrId, setWetvrId] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    setJavrScraper(state?.config?.scraper_settings?.javr?.javrScraper ?? 'javdatabase')
    setTpdbToken(state?.config?.vendor?.tpdb?.apiToken ?? '')
  }, [state])

  const { data: sites } = useQuery({ queryKey: ['sites'], queryFn: () => api.get<Site[]>('/options/sites') })

  const scrapeJavr = async () => {
    if (!javrId.trim()) return
    await api.post('/task/scrape-javr', { s: javrScraper, q: javrId.trim() })
    toast.info('JAVR scrape started')
  }

  const scrapeTpdb = async () => {
    if (!tpdbUrl.trim() || !tpdbToken.trim()) return
    await api.post('/task/scrape-tpdb', { apiToken: tpdbToken, sceneUrl: tpdbUrl })
    toast.info('TPDB scrape started')
  }

  const createCustom = async (andEdit: boolean) => {
    if (!customTitle.trim()) return
    const scene = await api.post<Scene>('/scene/create', { title: customTitle.trim(), id: customId.trim() })
    toast.success('Scene created')
    navigate(`/scenes/${scene.scene_id}${andEdit ? '?edit=1' : ''}`)
  }

  const scrapeByUrl = async () => {
    const url = scrapeUrl.trim()
    if (!url) return
    const lower = url.toLowerCase()

    let site = SINGLE_SCENE_SITES.find((s) => lower.includes(s.match))?.site ?? ''
    // also match against known scraper ids/domains
    if (!site) {
      for (const s of sites ?? []) {
        if (s.id && lower.includes(s.id)) {
          site = s.id
          break
        }
      }
    }
    if (!site) {
      toast.error('No scrapers exist for this domain')
      return
    }
    if (lower.includes('wetvr') && !wetvrId.trim()) {
      toast.error('WetVR scenes require the scene id')
      return
    }
    setBusy(true)
    try {
      const res = await api.post<{ status: string; scene?: Scene }>('/task/singlescrape', {
        site,
        sceneurl: url,
        additionalinfo: wetvrId.trim() ? [wetvrId.trim()] : []
      })
      if (res?.scene?.scene_id) {
        toast.success('Scene scraped')
        navigate(`/scenes/${res.scene.scene_id}?edit=1`)
      } else {
        toast.error(res?.status || 'Scrape failed')
      }
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="space-y-4">
      <SectionCard title="Import JAVR scene">
        <div className="flex flex-wrap items-end gap-2">
          <Field label="Scraper">
            <select value={javrScraper} onChange={(e) => setJavrScraper(e.target.value)} className={inputCls}>
              {JAVR_SCRAPERS.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Content ID">
            <input value={javrId} onChange={(e) => setJavrId(e.target.value)} className={`${inputCls} w-48`} placeholder="ABC-123" />
          </Field>
          <button onClick={scrapeJavr} disabled={!javrId.trim()} className={btnPrimaryCls}>
            Scrape JAVR
          </button>
        </div>
      </SectionCard>

      <SectionCard title="Import TPDB scene">
        <div className="flex flex-wrap items-end gap-2">
          <Field label="API token">
            <input
              type="password"
              value={tpdbToken}
              onChange={(e) => setTpdbToken(e.target.value)}
              className={`${inputCls} w-56`}
            />
          </Field>
          <Field label="Scene URL">
            <input value={tpdbUrl} onChange={(e) => setTpdbUrl(e.target.value)} className={`${inputCls} w-72 font-mono text-xs`} />
          </Field>
          <button onClick={scrapeTpdb} disabled={!tpdbUrl.trim() || !tpdbToken.trim()} className={btnPrimaryCls}>
            Scrape TPDB
          </button>
        </div>
      </SectionCard>

      <SectionCard title="Create custom scene">
        <div className="flex flex-wrap items-end gap-2">
          <Field label="Scene Id (optional)" hint="cannot be changed later">
            <input value={customId} onChange={(e) => setCustomId(e.target.value)} className={`${inputCls} w-48 font-mono text-xs`} />
          </Field>
          <Field label="Title">
            <input value={customTitle} onChange={(e) => setCustomTitle(e.target.value)} className={`${inputCls} w-72`} />
          </Field>
          <button onClick={() => createCustom(false)} disabled={!customTitle.trim()} className={btnPrimaryCls}>
            Create
          </button>
          <button onClick={() => createCustom(true)} disabled={!customTitle.trim()} className={btnPrimaryCls}>
            Create and edit
          </button>
        </div>
      </SectionCard>

      <SectionCard title="Scrape a scene by URL">
        <div className="flex flex-wrap items-end gap-2">
          <Field label="Scene URL">
            <input
              value={scrapeUrl}
              onChange={(e) => setScrapeUrl(e.target.value)}
              className={`${inputCls} w-96 font-mono text-xs`}
              placeholder="https://…"
            />
          </Field>
          {scrapeUrl.toLowerCase().includes('wetvr') && (
            <Field label="WetVR scene id">
              <input value={wetvrId} onChange={(e) => setWetvrId(e.target.value)} className={`${inputCls} w-32`} />
            </Field>
          )}
          <button onClick={scrapeByUrl} disabled={!scrapeUrl.trim() || busy} className={btnPrimaryCls}>
            {busy ? 'Scraping…' : 'Scrape scene'}
          </button>
        </div>
      </SectionCard>
    </div>
  )
}

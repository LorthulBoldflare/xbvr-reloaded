import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../../api/client'
import type { Playlist } from '../../../api/types'
import { useToastStore } from '../../../store/toasts'
import { useUIStore } from '../../../store/ui'
import { Toggle } from '../../../components/Toggle'
import { SectionCard, Field, btnCls, btnPrimaryCls, inputCls } from '../common'

// Content bundle backup/restore (old UI: Data import/export).
export function ImportExportSection() {
  const toast = useToastStore.getState()
  const askConfirm = useUIStore((s) => s.askConfirm)

  const [tab, setTab] = useState<'import' | 'export'>('export')
  const [incl, setIncl] = useState({
    inclScenes: true,
    inclCuepoints: true,
    inclHistory: true,
    inclActions: true,
    inclLinks: true,
    inclActors: true,
    inclActorAkas: true,
    inclActorActions: true,
    inclTagGroups: true,
    inclPlaylists: true,
    inclVolumes: true,
    inclSites: true,
    inclExtRefs: true,
    inclConfig: false
  })
  const [allSites, setAllSites] = useState(true)
  const [officialOnly, setOfficialOnly] = useState(false)
  const [playlistId, setPlaylistId] = useState(0)
  const [extRefSubset, setExtRefSubset] = useState('all')
  const [overwrite, setOverwrite] = useState(false)
  const [password, setPassword] = useState('')
  const [bundleUrl, setBundleUrl] = useState('')
  const [uploadData, setUploadData] = useState('')
  const [busy, setBusy] = useState(false)

  const { data: playlists } = useQuery({ queryKey: ['playlists', 'scene'], queryFn: () => api.get<Playlist[]>('/playlist') })

  const toggle = (k: keyof typeof incl) => setIncl((s) => ({ ...s, [k]: !s[k] }))

  const exportBundle = () => {
    if (!password) {
      toast.error('A bundle password is required')
      return
    }
    const params = new URLSearchParams({
      ...Object.fromEntries(Object.entries(incl).map(([k, v]) => [k, String(v)])),
      allSites: String(allSites),
      onlyIncludeOfficalSites: String(officialOnly),
      playlistId: String(playlistId),
      extRefSubset,
      download: 'true',
      bundlePassword: password
    })
    // fire the export, then download the staged bundle
    api.get(`/task/bundle/backup?${params}`, { toastOnError: true }).then(() => {
      window.location.href = '/download/xbvr-content-bundle.json'
    })
  }

  const importBundle = async () => {
    if (!password) {
      toast.error('A bundle password is required')
      return
    }
    if (!uploadData && !bundleUrl.trim()) {
      toast.error('Choose a bundle file or URL')
      return
    }
    const ok = await askConfirm({
      title: 'Restore content bundle?',
      message: overwrite ? 'Existing data will be OVERWRITTEN.' : 'Existing data will be kept where possible.',
      danger: overwrite
    })
    if (!ok) return
    setBusy(true)
    try {
      await api.post('/task/bundle/restore', {
        ...incl,
        allSites,
        onlyIncludeOfficalSites: officialOnly,
        extRefSubset,
        overwrite,
        uploadData,
        bundleUrl: bundleUrl.trim(),
        bundlePassword: password
      })
      toast.info('Restore started in the background')
    } finally {
      setBusy(false)
    }
  }

  const onFile = (f: File | undefined) => {
    if (!f) return
    const reader = new FileReader()
    reader.onload = () => setUploadData(String(reader.result))
    reader.readAsText(f)
  }

  const INCL_LABELS: [keyof typeof incl, string][] = [
    ['inclScenes', 'Scene data'],
    ['inclCuepoints', 'Cuepoints'],
    ['inclHistory', 'Watch history'],
    ['inclActions', 'Scene edits'],
    ['inclLinks', 'Matched files'],
    ['inclActors', 'Actors'],
    ['inclActorAkas', 'Actor aka groups'],
    ['inclActorActions', 'Actor edits'],
    ['inclTagGroups', 'Tag groups'],
    ['inclPlaylists', 'Saved searches'],
    ['inclVolumes', 'Storage paths'],
    ['inclSites', 'Scraper settings'],
    ['inclExtRefs', 'External references'],
    ['inclConfig', 'Config settings']
  ]

  return (
    <SectionCard
      title="Content bundle"
      actions={
        <div className="flex gap-1">
          {(['export', 'import'] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`rounded-lg px-3 py-1 text-sm ${tab === t ? 'bg-accent-soft font-semibold text-accent-strong' : 'text-muted'}`}
            >
              {t === 'export' ? 'Export' : 'Import'}
            </button>
          ))}
        </div>
      }
    >
      <div className="mb-4 grid max-w-2xl grid-cols-2 gap-x-6 gap-y-1.5 md:grid-cols-3">
        {INCL_LABELS.map(([k, label]) => (
          <Toggle key={k} checked={incl[k]} onChange={() => toggle(k)} label={label} />
        ))}
      </div>

      <div className="mb-4 flex max-w-2xl flex-wrap items-end gap-3">
        <Toggle checked={allSites} onChange={setAllSites} label="All studios" />
        <Toggle checked={officialOnly} onChange={setOfficialOnly} label="Only official studios" />
        {tab === 'export' && (
          <Field label="Filter by saved search">
            <select value={playlistId} onChange={(e) => setPlaylistId(Number(e.target.value))} className={inputCls}>
              <option value={0}>—</option>
              {(playlists ?? []).map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </Field>
        )}
        {incl.inclExtRefs && (
          <Field label="External references">
            <select value={extRefSubset} onChange={(e) => setExtRefSubset(e.target.value)} className={inputCls}>
              <option value="all">All</option>
              <option value="manual_matched">Manual matches</option>
              <option value="deleted_match">Deleted matches</option>
            </select>
          </Field>
        )}
        <Field label="Bundle password" hint="Required — encrypts credentials in the bundle">
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} className={`${inputCls} w-56`} />
        </Field>
      </div>

      {tab === 'import' && (
        <div className="mb-4 max-w-2xl space-y-3">
          <Field label="Bundle file">
            <input type="file" accept=".json" onChange={(e) => onFile(e.target.files?.[0])} className="text-sm" />
            {uploadData && <span className="mt-1 block text-xs text-ok">Bundle loaded ({(uploadData.length / 1024).toFixed(0)} KB)</span>}
          </Field>
          <Field label="…or bundle URL">
            <input value={bundleUrl} onChange={(e) => setBundleUrl(e.target.value)} className={`${inputCls} font-mono text-xs`} />
          </Field>
          <Toggle checked={overwrite} onChange={setOverwrite} label="Overwrite existing data" />
        </div>
      )}

      <div className="flex gap-2">
        {tab === 'export' ? (
          <button onClick={exportBundle} className={btnPrimaryCls}>
            Export bundle
          </button>
        ) : (
          <button onClick={importBundle} disabled={busy} className={btnPrimaryCls}>
            {busy ? 'Starting…' : 'Import bundle'}
          </button>
        )}
      </div>
    </SectionCard>
  )
}

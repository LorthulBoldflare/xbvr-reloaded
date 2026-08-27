import { lazy, Suspense, type ComponentType } from 'react'
import { Link, useParams } from 'react-router-dom'

const lazySection = <T extends Record<string, ComponentType>>(loader: () => Promise<T>, name: keyof T) =>
  lazy(() => loader().then((m) => ({ default: m[name] })))

const FilesSection = lazySection(() => import('./sections/FilesSection'), 'FilesSection')
const StorageSection = lazySection(() => import('./sections/StorageSection'), 'StorageSection')
const PreviewsSection = lazySection(() => import('./sections/PreviewsSection'), 'PreviewsSection')
const CacheSection = lazySection(() => import('./sections/CacheSection'), 'CacheSection')
const SchedulesSection = lazySection(() => import('./sections/SchedulesSection'), 'SchedulesSection')
const ScrapersSection = lazySection(() => import('./sections/ScrapersSection'), 'ScrapersSection')
const SceneCreateSection = lazySection(() => import('./sections/SceneCreateSection'), 'SceneCreateSection')
const FunscriptsSection = lazySection(() => import('./sections/FunscriptsSection'), 'FunscriptsSection')
const ImportExportSection = lazySection(() => import('./sections/ImportExportSection'), 'ImportExportSection')
const AuthenticationSection = lazySection(() => import('./sections/AuthenticationSection'), 'AuthenticationSection')
const PlayersSection = lazySection(() => import('./sections/PlayersSection'), 'PlayersSection')
const DlnaSection = lazySection(() => import('./sections/DlnaSection'), 'DlnaSection')
const WebUiSection = lazySection(() => import('./sections/WebUiSection'), 'WebUiSection')
const AdvancedSection = lazySection(() => import('./sections/AdvancedSection'), 'AdvancedSection')

const GROUPS: { name: string; sections: { id: string; label: string; component: () => JSX.Element }[] }[] = [
  {
    name: 'Options',
    sections: [
      { id: 'files', label: 'Files', component: () => <FilesSection /> },
      { id: 'storage', label: 'Storage', component: () => <StorageSection /> },
      { id: 'previews', label: 'Previews', component: () => <PreviewsSection /> },
      { id: 'cache', label: 'Cache', component: () => <CacheSection /> },
      { id: 'schedules', label: 'Task Schedules', component: () => <SchedulesSection /> }
    ]
  },
  {
    name: 'Scene data',
    sections: [
      { id: 'scrapers', label: 'Scrapers', component: () => <ScrapersSection /> },
      { id: 'create', label: 'Create / Import scene', component: () => <SceneCreateSection /> },
      { id: 'funscripts', label: 'Funscripts', component: () => <FunscriptsSection /> },
      { id: 'bundle', label: 'Data import/export', component: () => <ImportExportSection /> }
    ]
  },
  {
    name: 'Interfaces',
    sections: [
      { id: 'authentication', label: 'Authentication', component: () => <AuthenticationSection /> },
      { id: 'players', label: 'Players', component: () => <PlayersSection /> },
      { id: 'dlna', label: 'DLNA', component: () => <DlnaSection /> },
      { id: 'web', label: 'Web UI', component: () => <WebUiSection /> },
      { id: 'advanced', label: 'Advanced', component: () => <AdvancedSection /> }
    ]
  }
]

export function OptionsPage() {
  const params = useParams()
  const section = params['*']?.split('/')[0] || 'files'
  const active = GROUPS.flatMap((g) => g.sections).find((s) => s.id === section) ?? GROUPS[0].sections[0]

  return (
    <div className="flex gap-8">
      <aside className="sticky top-5 w-44 shrink-0 self-start">
        {GROUPS.map((g) => (
          <div key={g.name} className="mb-5">
            <div className="mb-1.5 px-2.5 text-[10px] font-bold uppercase tracking-widest text-muted">{g.name}</div>
            {g.sections.map((s) => (
              <Link
                key={s.id}
                to={`/options/${s.id}`}
                className={`block rounded-xl px-2.5 py-1.5 text-sm transition-colors ${
                  s.id === active.id
                    ? 'bg-accent-soft font-semibold text-accent-strong shadow-[inset_2px_0_0_0_var(--accent)]'
                    : 'text-muted hover:bg-surface-2 hover:text-fg'
                }`}
              >
                {s.label}
              </Link>
            ))}
          </div>
        ))}
      </aside>
      <div className="min-w-0 max-w-5xl flex-1">
        <Suspense fallback={<div className="py-16 text-center text-muted">Loading…</div>}>
          <active.component />
        </Suspense>
      </div>
    </div>
  )
}

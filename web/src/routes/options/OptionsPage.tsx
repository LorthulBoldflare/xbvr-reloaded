import { useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { FilesSection } from './sections/FilesSection'
import { StorageSection } from './sections/StorageSection'
import { PreviewsSection } from './sections/PreviewsSection'
import { CacheSection } from './sections/CacheSection'
import { SchedulesSection } from './sections/SchedulesSection'
import { ScrapersSection } from './sections/ScrapersSection'
import { SceneCreateSection } from './sections/SceneCreateSection'
import { FunscriptsSection } from './sections/FunscriptsSection'
import { ImportExportSection } from './sections/ImportExportSection'
import { AuthenticationSection } from './sections/AuthenticationSection'
import { PlayersSection } from './sections/PlayersSection'
import { DlnaSection } from './sections/DlnaSection'
import { WebUiSection } from './sections/WebUiSection'
import { AdvancedSection } from './sections/AdvancedSection'

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
    <div className="flex gap-6">
      <aside className="sticky top-14 w-44 shrink-0 self-start">
        {GROUPS.map((g) => (
          <div key={g.name} className="mb-4">
            <div className="mb-1 px-2 text-xs font-semibold uppercase tracking-wide text-muted">{g.name}</div>
            {g.sections.map((s) => (
              <Link
                key={s.id}
                to={`/options/${s.id}`}
                className={`block rounded-lg px-2 py-1.5 text-sm ${
                  s.id === active.id ? 'bg-accent-soft font-semibold text-accent-strong' : 'text-muted hover:bg-surface-2 hover:text-fg'
                }`}
              >
                {s.label}
              </Link>
            ))}
          </div>
        ))}
      </aside>
      <div className="min-w-0 flex-1">
        <active.component />
      </div>
    </div>
  )
}

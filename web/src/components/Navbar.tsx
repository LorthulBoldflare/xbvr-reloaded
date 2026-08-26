import { useState } from 'react'
import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import { getTheme, toggleTheme } from '../theme'
import { useMessagesStore } from '../store/messages'
import { useUIStore } from '../store/ui'
import { useToastStore } from '../store/toasts'
import { useOptionsStorage, useVersionCheck } from '../api/hooks'
import { Popover } from '../components/Popover'
import { MoonIcon, SunIcon } from '../components/icons'

function StatusChip({ label, active, message }: { label: string; active: boolean; message: string }) {
  return (
    <span
      title={message || label}
      className="flex items-center gap-1.5 rounded-full border border-line bg-surface px-2.5 py-1 text-xs text-muted"
    >
      <span className={`h-1.5 w-1.5 rounded-full ${active ? 'animate-pulse bg-ok' : 'bg-line-strong'}`} />
      {label}
    </span>
  )
}

function ActionsMenu() {
  const { data: storage } = useOptionsStorage()
  const queryClient = useQueryClient()
  const trigger = useMutation({
    mutationFn: (name: string) => api.get(`/task/webhook/${name}`),
    onSuccess: () => useToastStore.getState().success('Webhook triggered'),
    onSettled: () => queryClient.invalidateQueries()
  })

  const hooks: { key: string; label: string; configured: boolean }[] = [
    {
      key: 'trigger-import',
      label: 'Trigger external import',
      configured: !!storage?.webhooks?.trigger_external_import?.url
    },
    {
      key: 'refresh-import',
      label: 'Refresh external import',
      configured: !!storage?.webhooks?.refresh_external_import?.url
    }
  ]
  const visible = hooks.filter((h) => h.configured)
  if (visible.length === 0) return null

  return (
    <Popover
      button={
        <>
          Actions <span className="text-muted">▾</span>
        </>
      }
      width="w-56"
      align="right"
    >
      {(close) => (
        <div className="flex flex-col">
          {visible.map((h) => (
            <button
              key={h.key}
              disabled={trigger.isPending}
              onClick={() => {
                trigger.mutate(h.key)
                close()
              }}
              className="rounded px-2 py-1.5 text-left text-sm hover:bg-accent-soft"
            >
              {h.label}
            </button>
          ))}
        </div>
      )}
    </Popover>
  )
}

export function Navbar() {
  const [theme, setThemeState] = useState(getTheme())
  const { data: version } = useVersionCheck()
  const { lockScrape, lockRescan, lastScrapeMessage, lastRescanMessage } = useMessagesStore()
  const openQuickFind = useUIStore((s) => s.openQuickFind)

  const navClass = ({ isActive }: { isActive: boolean }) =>
    `rounded-lg px-3 py-1.5 text-sm font-medium ${isActive ? 'bg-accent-soft text-accent-strong' : 'text-muted hover:text-fg'}`

  return (
    <>
      <header className="sticky top-0 z-30 border-b border-line bg-surface">
        <div className="mx-auto flex max-w-[1600px] items-center gap-2 px-4 py-2">
          <Link to="/" className="mr-2 flex items-center gap-2 text-lg font-bold tracking-tight">
            XBVR
            {version?.current_version && (
              <span className="rounded-full bg-surface-2 px-2 py-0.5 text-[10px] font-normal text-muted">
                {version.current_version}
              </span>
            )}
          </Link>
          <nav className="flex items-center gap-1">
            <NavLink to="/" end className={navClass}>
              Scenes
            </NavLink>
            <NavLink to="/actors" className={navClass}>
              Actors
            </NavLink>
            <NavLink to="/options" className={navClass}>
              Options
            </NavLink>
          </nav>
          <button
            onClick={() => openQuickFind(false)}
            className="ml-2 flex items-center gap-2 rounded-lg border border-line bg-surface-2 px-3 py-1.5 text-sm text-muted hover:border-line-strong"
            title="Quick find (?)"
          >
            Quick find…
            <kbd className="rounded border border-line bg-surface px-1 text-[10px]">?</kbd>
          </button>
          <div className="ml-auto flex items-center gap-2">
            <StatusChip label="Files" active={lockRescan} message={lastRescanMessage} />
            <StatusChip label="Data" active={lockScrape} message={lastScrapeMessage} />
            <ActionsMenu />
            <button
              onClick={() => setThemeState(toggleTheme())}
              className="rounded-lg border border-line bg-surface px-2.5 py-1.5 text-sm hover:border-line-strong"
              title="Toggle light/dark theme"
            >
              {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
            </button>
          </div>
        </div>
      </header>
      {version?.update_notify && version.latest_version && version.latest_version !== version.current_version && (
        <UpdateSnackbar version={version.latest_version} />
      )}
    </>
  )
}

function UpdateSnackbar({ version }: { version: string }) {
  const [dismissed, setDismissed] = useState(false)
  if (dismissed) return null
  return (
    <div className="fixed bottom-4 left-1/2 z-40 flex -translate-x-1/2 items-center gap-3 rounded-lg border border-warn/40 bg-surface px-4 py-2 shadow-xl">
      <span className="text-sm">
        Version <strong>{version}</strong> available!
      </span>
      <a
        href="https://github.com/xbapps/xbvr/releases"
        target="_blank"
        rel="noreferrer"
        className="rounded-lg bg-accent px-3 py-1 text-sm text-white"
      >
        Download now
      </a>
      <button onClick={() => setDismissed(true)} className="text-muted hover:text-fg" aria-label="Dismiss">
        ✕
      </button>
    </div>
  )
}

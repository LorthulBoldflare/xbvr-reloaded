import { useState } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import { getTheme, toggleTheme } from '../theme'
import { useMessagesStore } from '../store/messages'
import { useUIStore } from '../store/ui'
import { useToastStore } from '../store/toasts'
import { useOptionsStorage, useVersionCheck } from '../api/hooks'
import { Popover } from '../components/Popover'
import { ChevronsLeftIcon, ChevronsRightIcon, FilmIcon, MoonIcon, SearchIcon, SlidersIcon, SunIcon, UsersIcon } from '../components/icons'

const SIDEBAR_KEY = 'xbvr-sidebar'

export function useSidebarCollapsed(): [boolean, () => void] {
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem(SIDEBAR_KEY) === 'collapsed')
  const toggle = () => {
    setCollapsed((c) => {
      localStorage.setItem(SIDEBAR_KEY, c ? 'expanded' : 'collapsed')
      return !c
    })
  }
  return [collapsed, toggle]
}

function StatusDot({ label, active, message, collapsed }: { label: string; active: boolean; message: string; collapsed: boolean }) {
  return (
    <span
      title={message || label}
      className="flex items-center gap-2 rounded-lg px-2 py-1 text-xs text-muted"
    >
      <span className={`h-1.5 w-1.5 rounded-full ${active ? 'animate-pulse bg-accent' : 'bg-line-strong'}`} />
      {!collapsed && <span>{label}</span>}
    </span>
  )
}

function ActionsMenu({ collapsed }: { collapsed: boolean }) {
  const { data: storage } = useOptionsStorage()
  const queryClient = useQueryClient()
  const trigger = useMutation({
    mutationFn: (name: string) => api.get(`/task/webhook/${name}`),
    onSuccess: () => useToastStore.getState().success('Webhook triggered'),
    onSettled: () => queryClient.invalidateQueries()
  })

  const hooks = [
    { key: 'trigger-import', label: 'Trigger external import', configured: !!storage?.webhooks?.trigger_external_import?.url },
    { key: 'refresh-import', label: 'Refresh external import', configured: !!storage?.webhooks?.refresh_external_import?.url }
  ].filter((h) => h.configured)
  if (hooks.length === 0) return null

  return (
    <Popover
      button={
        <span className="flex w-full items-center gap-2">
          <SlidersIcon className="h-4 w-4" />
          {!collapsed && <span>Actions</span>}
          {!collapsed && <span className="ml-auto text-muted">▾</span>}
        </span>
      }
      width="w-56"
      buttonClassName="w-full rounded-xl px-3 py-2 text-sm text-muted hover:bg-surface-2 hover:text-fg"
    >
      {(close) => (
        <div className="flex flex-col">
          {hooks.map((h) => (
            <button
              key={h.key}
              disabled={trigger.isPending}
              onClick={() => {
                trigger.mutate(h.key)
                close()
              }}
              className="rounded-lg px-2 py-1.5 text-left text-sm hover:bg-accent-soft"
            >
              {h.label}
            </button>
          ))}
        </div>
      )}
    </Popover>
  )
}

// Left navigation rail — collapsible (persisted in localStorage).
export function Sidebar() {
  const [theme, setThemeState] = useState(getTheme())
  const [collapsed, toggleCollapsed] = useSidebarCollapsed()
  const { data: version } = useVersionCheck()
  const lockScrape = useMessagesStore((s) => s.lockScrape)
  const lockRescan = useMessagesStore((s) => s.lockRescan)
  const lastScrapeMessage = useMessagesStore((s) => s.lastScrapeMessage)
  const lastRescanMessage = useMessagesStore((s) => s.lastRescanMessage)
  const openQuickFind = useUIStore((s) => s.openQuickFind)

  const navItems = [
    { to: '/', label: 'Scenes', icon: FilmIcon, end: true },
    { to: '/actors', label: 'Actors', icon: UsersIcon, end: false },
    { to: '/options', label: 'Options', icon: SlidersIcon, end: false }
  ]

  return (
    <aside
      className={`sticky top-0 flex h-screen shrink-0 flex-col border-r border-line bg-surface transition-[width] duration-200 ${
        collapsed ? 'w-16' : 'w-56'
      }`}
    >
      {/* brand */}
      <Link to="/" className="flex items-center gap-2.5 px-3 py-4" title="XBVR home">
        <span className="brand-gradient flex h-8 w-8 shrink-0 items-center justify-center rounded-lg font-black text-white">
          X
        </span>
        {!collapsed && (
          <span className="flex min-w-0 flex-col">
            <span className="text-base font-extrabold leading-none tracking-tight">XBVR</span>
            {version?.current_version && (
              <span className="mt-0.5 text-[10px] text-muted">{version.current_version}</span>
            )}
          </span>
        )}
      </Link>

      {/* quick find */}
      <div className="px-2">
        <button
          onClick={() => openQuickFind(false)}
          className="flex w-full items-center gap-2 rounded-xl border border-line bg-surface-2 px-2.5 py-2 text-sm text-muted hover:border-line-strong hover:text-fg"
          title="Quick find (?)"
        >
          <SearchIcon className="h-4 w-4 shrink-0" />
          {!collapsed && <span className="flex-1 text-left">Quick find</span>}
          {!collapsed && <kbd className="rounded border border-line bg-surface px-1 text-[10px]">?</kbd>}
        </button>
      </div>

      {/* nav */}
      <nav className="mt-3 flex flex-col gap-0.5 px-2">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            title={item.label}
            className={({ isActive }) =>
              `flex items-center gap-2.5 rounded-xl px-3 py-2 text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-accent-soft text-accent-strong shadow-[inset_2px_0_0_0_var(--accent)]'
                  : 'text-muted hover:bg-surface-2 hover:text-fg'
              }`
            }
          >
            <item.icon className="h-4 w-4 shrink-0" />
            {!collapsed && <span>{item.label}</span>}
          </NavLink>
        ))}
      </nav>

      <div className="flex-1" />

      {/* status + actions + theme + collapse */}
      <div className="flex flex-col gap-0.5 border-t border-line px-2 py-2">
        <StatusDot label="File scan" active={lockRescan} message={lastRescanMessage} collapsed={collapsed} />
        <StatusDot label="Scraping" active={lockScrape} message={lastScrapeMessage} collapsed={collapsed} />
        <ActionsMenu collapsed={collapsed} />
        <button
          onClick={() => setThemeState(toggleTheme())}
          className="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-sm text-muted hover:bg-surface-2 hover:text-fg"
          title="Toggle light/dark theme"
        >
          {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
          {!collapsed && <span>{theme === 'dark' ? 'Light theme' : 'Dark theme'}</span>}
        </button>
        <button
          onClick={toggleCollapsed}
          className="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-sm text-muted hover:bg-surface-2 hover:text-fg"
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <ChevronsRightIcon /> : <ChevronsLeftIcon />}
          {!collapsed && <span>Collapse</span>}
        </button>
      </div>
    </aside>
  )
}

export function UpdateSnackbar() {
  const { data: version } = useVersionCheck()
  const [dismissed, setDismissed] = useState(false)
  if (!version?.update_notify || !version.latest_version || version.latest_version === version.current_version || dismissed)
    return null
  return (
    <div className="fixed bottom-4 left-1/2 z-40 flex -translate-x-1/2 items-center gap-3 rounded-full border border-warn/40 bg-surface px-4 py-2 shadow-2xl">
      <span className="text-sm">
        Version <strong>{version.latest_version}</strong> available
      </span>
      <a
        href="https://github.com/xbapps/xbvr/releases"
        target="_blank"
        rel="noreferrer"
        className="rounded-full bg-accent px-3 py-1 text-sm font-semibold text-accent-fg"
      >
        Download
      </a>
      <button onClick={() => setDismissed(true)} className="text-muted hover:text-fg" aria-label="Dismiss">
        ✕
      </button>
    </div>
  )
}

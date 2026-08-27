import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api/client'
import type { ResponseGetFilters } from '../../api/types'
import { useSceneFilterStore, DL_STATES, applyDlState, normalizeSceneFilters } from '../../store/sceneFilters'
import { useToastStore } from '../../store/toasts'
import { useUIStore } from '../../store/ui'
import { akaApi, tagGroupApi } from '../../api/groups'
import { TagFilterInput } from '../../components/TagFilterInput'
import { SavedSearchPicker } from '../../components/SavedSearchPicker'
import { Toggle } from '../../components/Toggle'
import { Modal } from '../../components/Modal'

export function useSceneFilterOptions() {
  return useQuery({
    queryKey: ['sceneFilters'],
    queryFn: () => api.get<ResponseGetFilters>('/scene/filters'),
    staleTime: 5 * 60_000
  })
}

const stripPrefix = (v: string) => (v.startsWith('!') || v.startsWith('&') ? v.slice(1) : v)

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="border-t border-line pt-3 first:border-t-0 first:pt-0">
      <div className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-muted">{title}</div>
      {children}
    </div>
  )
}

// The scenes Filters popover content (saved searches, availability,
// properties, all tag inputs, group management).
export function FiltersPopoverContent({ counts }: { counts?: Record<string, number> }) {
  const { filters, patch, setFilters } = useSceneFilterStore()
  const { data: opts } = useSceneFilterOptions()
  const queryClient = useQueryClient()
  const askConfirm = useUIStore((s) => s.askConfirm)
  const toast = useToastStore.getState()

  const [tagGroupDialog, setTagGroupDialog] = useState<'create' | 'rename' | null>(null)
  const [tagGroupName, setTagGroupName] = useState('')

  const reloadFilters = () => queryClient.invalidateQueries({ queryKey: ['sceneFilters'] })
  const warn = (status: string) => status && toast.info(`Warning: ${status}`)

  const plainCast = filters.cast.filter((c) => !stripPrefix(c).startsWith('aka:'))
  const akaSelected = filters.cast.find((c) => stripPrefix(c).startsWith('aka:'))

  const plainTags = filters.tags.filter((t) => !stripPrefix(t).startsWith('tag group:'))
  const tagGroupSelected = filters.tags.find((t) => stripPrefix(t).startsWith('tag group:'))

  const properties: { key: string; label: string }[] = [
    { key: 'watchlist', label: 'Watchlist' },
    { key: 'favourite', label: 'Favourite' },
    { key: 'wishlist', label: 'Wishlist' },
    { key: 'scripted', label: 'Scripted' }
  ]

  const countsFor: Record<string, number | undefined> = {
    any: counts?.count_any,
    available: counts?.count_available,
    downloaded: counts?.count_downloaded,
    missing: counts?.count_not_downloaded,
    hidden: counts?.count_hidden
  }

  return (
    <div className="space-y-3">
      <SavedSearchPicker
        type="scene"
        currentFilters={filters as unknown as Record<string, unknown>}
        onApply={(f) => {
          // Saved searches come from the server and may be legacy payloads —
          // normalize before they enter the store (null arrays etc.).
          const normalized = normalizeSceneFilters(f as never)
          setFilters(applyDlState(normalized, normalized.dlState))
        }}
      />

      <Section title="Availability">
        <div className="flex flex-col gap-1">
          {DL_STATES.map((d) => (
            <button
              key={d.value}
              onClick={() => useSceneFilterStore.setState({ filters: applyDlState(filters, d.value) })}
              className={`flex justify-between rounded-lg px-2 py-1 text-left text-sm ${
                filters.dlState === d.value ? 'bg-accent-soft font-semibold text-accent-strong' : 'hover:bg-surface-2'
              }`}
            >
              <span>{d.label}</span>
              {countsFor[d.value] !== undefined && <span className="text-muted">{countsFor[d.value]}</span>}
            </button>
          ))}
        </div>
      </Section>

      <Section title="Properties">
        <div className="flex flex-wrap gap-1.5">
          {properties.map((p) => (
            <button
              key={p.key}
              onClick={() =>
                patch({
                  lists: filters.lists.includes(p.key)
                    ? filters.lists.filter((l) => l !== p.key)
                    : [...filters.lists, p.key]
                })
              }
              className={`rounded-full border px-2.5 py-1 text-xs font-medium ${
                filters.lists.includes(p.key)
                  ? 'border-accent bg-accent-soft text-accent-strong'
                  : 'border-line text-muted hover:text-fg'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </Section>

      <Section title="Watched">
        <div className="flex gap-1">
          {[
            { v: null, label: 'Everything' },
            { v: true, label: 'Watched' },
            { v: false, label: 'Unwatched' }
          ].map((o) => (
            <button
              key={o.label}
              onClick={() => patch({ isWatched: o.v as boolean | null })}
              className={`rounded-lg px-2 py-1 text-xs ${
                filters.isWatched === o.v ? 'bg-accent-soft font-semibold text-accent-strong' : 'text-muted hover:bg-surface-2'
              }`}
            >
              {o.label}
            </button>
          ))}
        </div>
      </Section>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Release month</div>
          <div className="flex gap-1">
            <select
              value={filters.releaseMonth}
              onChange={(e) => patch({ releaseMonth: e.target.value })}
              className="min-w-0 flex-1 rounded-lg border border-line bg-surface-2 px-2 py-1 text-sm"
            >
              <option value="">—</option>
              {(opts?.release_month ?? []).map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
            {filters.releaseMonth && (
              <button onClick={() => patch({ releaseMonth: '' })} className="rounded-lg border border-line px-2 text-sm">
                ✕
              </button>
            )}
          </div>
        </div>
        <div>
          <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Folder</div>
          <div className="flex gap-1">
            <select
              value={filters.volume}
              onChange={(e) => {
                patch({ volume: Number(e.target.value) })
                reloadFilters()
              }}
              className="min-w-0 flex-1 rounded-lg border border-line bg-surface-2 px-2 py-1 text-sm"
            >
              <option value={0}>—</option>
              {(opts?.volumes ?? []).map((v) => (
                <option key={v.id} value={v.id}>
                  {v.path}
                </option>
              ))}
            </select>
            {filters.volume !== 0 && (
              <button
                onClick={() => {
                  patch({ volume: 0 })
                  reloadFilters()
                }}
                className="rounded-lg border border-line px-2 text-sm"
              >
                ✕
              </button>
            )}
          </div>
        </div>
      </div>

      <TagFilterInput label="Cast" values={filters.cast} options={opts?.cast ?? []} onChange={(v) => patch({ cast: v })} mode="3way" />
      <TagFilterInput label="Site" values={filters.sites} options={opts?.sites ?? []} onChange={(v) => patch({ sites: v })} mode="2way" />
      <TagFilterInput label="Tags" values={filters.tags} options={opts?.tags ?? []} onChange={(v) => patch({ tags: v })} mode="3way" />
      <TagFilterInput
        label="Cuepoints"
        values={filters.cuepoint}
        options={opts?.cuepoints ?? []}
        onChange={(v) => patch({ cuepoint: v })}
        mode="3way"
      />
      <TagFilterInput
        label="Attributes"
        values={filters.attributes}
        options={opts?.attributes ?? []}
        onChange={(v) => patch({ attributes: v })}
        mode="3way"
      />

      <Section title="Files">
        <Toggle
          checked={!!filters.showUnmatched}
          onChange={(v) => patch({ showUnmatched: v })}
          label="Show unmatched files in the grid"
        />
      </Section>

      <Section title="Actor groups (AKA)">
        <div className="flex flex-wrap gap-1.5">
          <button
            disabled={plainCast.length < 2}
            title="Create a group from the selected actors"
            onClick={async () => {
              const data = await akaApi.create(filters.cast)
              patch({ cast: [...filters.cast, data.akas.aka_actor.name] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            New
          </button>
          <button
            disabled={!akaSelected}
            title="Delete the selected aka group"
            onClick={async () => {
              if (!(await askConfirm({ title: `Delete aka group ${stripPrefix(akaSelected!)}?`, danger: true }))) return
              const data = await akaApi.delete(stripPrefix(akaSelected!))
              patch({ cast: [] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Delete
          </button>
          <button
            disabled={!akaSelected || plainCast.length === 0}
            title="Add selected actors to the group"
            onClick={async () => {
              const data = await akaApi.add(filters.cast)
              patch({ cast: [...plainCast, data.akas.aka_actor.name] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Add cast
          </button>
          <button
            disabled={!akaSelected || plainCast.length === 0}
            title="Remove selected actors from the group"
            onClick={async () => {
              const data = await akaApi.remove(filters.cast)
              patch({ cast: [...plainCast, data.akas.aka_actor.name] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Remove cast
          </button>
        </div>
      </Section>

      <Section title="Tag groups">
        <div className="flex flex-wrap gap-1.5">
          <button
            disabled={plainTags.length < 2}
            onClick={() => {
              setTagGroupName('')
              setTagGroupDialog('create')
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            New
          </button>
          <button
            disabled={!tagGroupSelected}
            onClick={async () => {
              if (!(await askConfirm({ title: `Delete tag group ${stripPrefix(tagGroupSelected!)}?`, danger: true }))) return
              const data = await tagGroupApi.delete(stripPrefix(tagGroupSelected!))
              patch({ tags: [] })
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Delete
          </button>
          <button
            disabled={!tagGroupSelected || plainTags.length === 0}
            onClick={async () => {
              const data = await tagGroupApi.add(filters.tags)
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Add tag
          </button>
          <button
            disabled={!tagGroupSelected || plainTags.length === 0}
            onClick={async () => {
              const data = await tagGroupApi.remove(filters.tags)
              reloadFilters()
              warn(data.status)
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Remove tag
          </button>
          <button
            disabled={!tagGroupSelected}
            onClick={() => {
              setTagGroupName('')
              setTagGroupDialog('rename')
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            Rename
          </button>
          <button
            disabled={!tagGroupSelected}
            title="Replace the filter with the group's tags"
            onClick={async () => {
              const data = await tagGroupApi.get(stripPrefix(tagGroupSelected!))
              if (data.status) {
                toast.error(data.status)
                return
              }
              patch({ tags: [`tag group:${data.tag_group.name}`, ...data.tag_group.tags.map((t) => t.name)] })
              reloadFilters()
            }}
            className="rounded-lg border border-line px-2 py-1 text-xs hover:bg-surface-2 disabled:opacity-40"
          >
            List tags
          </button>
        </div>
      </Section>

      <Modal
        open={tagGroupDialog !== null}
        onClose={() => setTagGroupDialog(null)}
        width="max-w-sm"
        title={tagGroupDialog === 'create' ? 'Create tag group' : 'Rename tag group'}
      >
        <input
          autoFocus
          value={tagGroupName}
          onChange={(e) => setTagGroupName(e.target.value)}
          placeholder="Group name"
          className="mb-4 w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm"
        />
        <div className="flex justify-end gap-2">
          <button onClick={() => setTagGroupDialog(null)} className="rounded-lg border border-line px-3 py-1.5 text-sm">
            Cancel
          </button>
          <button
            disabled={!tagGroupName.trim()}
            onClick={async () => {
              const action = tagGroupDialog
              setTagGroupDialog(null)
              if (action === 'create') {
                const data = await tagGroupApi.create(tagGroupName, filters.tags)
                if (data.tag_group.tag_group_tag.name) {
                  patch({ tags: [...filters.tags, data.tag_group.tag_group_tag.name] })
                }
                reloadFilters()
                warn(data.status)
              } else {
                const data = await tagGroupApi.rename(tagGroupName, filters.tags)
                if (data.status) {
                  toast.error(data.status)
                } else {
                  patch({ tags: [data.tag_group.tag_group_tag.name] })
                  reloadFilters()
                }
              }
            }}
            className="rounded-lg bg-accent px-3 py-1.5 text-sm text-white disabled:opacity-50"
          >
            Save
          </button>
        </div>
      </Modal>
    </div>
  )
}

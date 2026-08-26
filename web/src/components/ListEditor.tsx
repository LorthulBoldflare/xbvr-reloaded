import { safeHref } from '../lib/format'
import { PlusIcon, TrashIcon } from './icons'

// Generic editable string-list (used for filenames, aliases, tattoos, links,
// image URLs in the edit forms).
export function ListEditor({
  items,
  onChange,
  showLinks = false,
  addLabel = 'Add'
}: {
  items: string[]
  onChange: (items: string[]) => void
  showLinks?: boolean
  addLabel?: string
}) {
  return (
    <div className="space-y-1">
      {items.map((item, i) => (
        <div key={i} className="flex items-center gap-1">
          <input
            value={item}
            onChange={(e) => onChange(items.map((x, j) => (j === i ? e.target.value : x)))}
            className="min-w-0 flex-1 rounded-lg border border-line bg-surface-2 px-2 py-1 font-mono text-xs"
          />
          {showLinks && (
            <a
              href={safeHref(item)}
              target="_blank"
              rel="noreferrer"
              className="rounded-lg border border-line px-2 py-1 text-xs text-muted hover:text-fg"
              title="Open link"
            >
              ↗
            </a>
          )}
          <button
            onClick={() => onChange(items.filter((_, j) => j !== i))}
            className="rounded-lg border border-line px-2 py-1 text-xs text-muted hover:text-danger"
            title="Delete row"
          >
            <TrashIcon />
          </button>
        </div>
      ))}
      <button
        onClick={() => onChange([...items, ''])}
        className="flex items-center gap-1 rounded-lg border border-line px-2 py-1 text-xs text-muted hover:text-fg"
      >
        <PlusIcon /> {addLabel}
      </button>
    </div>
  )
}

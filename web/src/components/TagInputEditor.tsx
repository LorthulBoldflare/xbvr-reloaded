import { useMemo, useState } from 'react'

// Chip editor with autocomplete and allow-new (used by scene/actor edit
// forms — unlike TagFilterInput, values are plain names and new entries are
// allowed).
export function TagInputEditor({
  label,
  values,
  options,
  onChange
}: {
  label: string
  values: string[]
  options: string[]
  onChange: (values: string[]) => void
}) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)

  const filtered = useMemo(() => {
    const q = query.toLowerCase()
    return options.filter((o) => !values.includes(o) && o.toLowerCase().includes(q)).slice(0, 30)
  }, [options, values, query])

  const add = (name: string) => {
    const v = name.trim()
    if (v && !values.includes(v)) onChange([...values, v])
    setQuery('')
  }

  return (
    <div>
      <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">{label}</div>
      <div className="flex flex-wrap gap-1 rounded-lg border border-line bg-surface-2 p-1.5">
        {values.map((v) => (
          <span key={v} className="flex items-center gap-1 rounded-full bg-surface-3 px-2 py-0.5 text-xs">
            {v}
            <button onClick={() => onChange(values.filter((x) => x !== v))} className="text-muted hover:text-danger">
              ✕
            </button>
          </span>
        ))}
        <div className="relative min-w-24 flex-1">
          <input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setOpen(true)
            }}
            onFocus={() => setOpen(true)}
            onBlur={() => setTimeout(() => setOpen(false), 150)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                if (filtered.length > 0) add(filtered[0])
                else add(query)
              }
            }}
            className="w-full bg-transparent px-1 py-0.5 text-sm outline-none"
          />
          {open && (filtered.length > 0 || query.trim()) && (
            <div className="absolute left-0 top-full z-30 mt-1 max-h-56 w-64 overflow-y-auto rounded-lg border border-line bg-surface shadow-xl">
              {filtered.map((o) => (
                <button
                  key={o}
                  onMouseDown={(e) => {
                    e.preventDefault()
                    add(o)
                  }}
                  className="block w-full truncate px-2 py-1 text-left text-sm hover:bg-accent-soft"
                >
                  {o}
                </button>
              ))}
              {query.trim() && !options.includes(query.trim()) && (
                <button
                  onMouseDown={(e) => {
                    e.preventDefault()
                    add(query)
                  }}
                  className="block w-full px-2 py-1 text-left text-sm italic text-muted hover:bg-accent-soft"
                >
                  Add "{query.trim()}"
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

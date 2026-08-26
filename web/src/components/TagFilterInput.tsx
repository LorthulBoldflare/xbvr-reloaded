import { useMemo, useRef, useState } from 'react'

export type ChipMode = '2way' | '3way'

function chipState(v: string): { name: string; mode: 'include' | 'must' | 'exclude' } {
  if (v.startsWith('!')) return { name: v.slice(1), mode: 'exclude' }
  if (v.startsWith('&')) return { name: v.slice(1), mode: 'must' }
  return { name: v, mode: 'include' }
}

// Multi-value autocomplete with the old UI's chip-cycling semantics:
// click a chip to cycle include → must (&) → exclude (!) → remove
// (2-way: include → exclude → remove).
export function TagFilterInput({
  label,
  values,
  options,
  onChange,
  mode = '3way',
  placeholder = 'Type to filter…'
}: {
  label: string
  values: string[]
  options: string[]
  onChange: (values: string[]) => void
  mode?: ChipMode
  placeholder?: string
}) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const selected = useMemo(() => new Set(values.map((v) => chipState(v).name)), [values])
  const filtered = useMemo(() => {
    const q = query.toLowerCase()
    return options.filter((o) => !selected.has(o) && o.toLowerCase().includes(q)).slice(0, 30)
  }, [options, selected, query])

  const cycle = (v: string) => {
    const { name, mode: m } = chipState(v)
    let next: string | null
    if (mode === '3way') {
      next = m === 'include' ? `&${name}` : m === 'must' ? `!${name}` : null
    } else {
      next = m === 'include' ? `!${name}` : null
    }
    onChange(next === null ? values.filter((x) => x !== v) : values.map((x) => (x === v ? next! : x)))
  }

  const add = (name: string) => {
    onChange([...values, name])
    setQuery('')
  }

  return (
    <div ref={ref}>
      <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">{label}</div>
      <div className="flex flex-wrap gap-1 rounded-lg border border-line bg-surface-2 p-1.5">
        {values.map((v) => {
          const { name, mode: m } = chipState(v)
          return (
            <button
              key={v}
              onClick={() => cycle(v)}
              title="Click to cycle: include → must-have → exclude → remove"
              className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                m === 'must'
                  ? 'bg-ok/20 text-ok'
                  : m === 'exclude'
                    ? 'bg-danger/20 text-danger'
                    : 'bg-surface-3 text-fg'
              }`}
            >
              {name}
            </button>
          )
        })}
        <div className="relative min-w-24 flex-1">
          <input
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setOpen(true)
            }}
            onFocus={() => setOpen(true)}
            onBlur={() => setTimeout(() => setOpen(false), 150)}
            placeholder={values.length === 0 ? placeholder : ''}
            className="w-full bg-transparent px-1 py-0.5 text-sm outline-none"
          />
          {open && filtered.length > 0 && (
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
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

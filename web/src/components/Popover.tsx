import { useEffect, useRef, useState, type ReactNode } from 'react'

// Popover: a trigger button + floating panel. Closes on Esc / outside click.
export function Popover({
  button,
  children,
  align = 'left',
  width = 'w-96',
  buttonClassName = ''
}: {
  button: ReactNode
  children: ReactNode | ((close: () => void) => ReactNode)
  align?: 'left' | 'right'
  width?: string
  buttonClassName?: string
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false)
    }
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    window.addEventListener('keydown', onKey, true)
    window.addEventListener('mousedown', onClick)
    return () => {
      window.removeEventListener('keydown', onKey, true)
      window.removeEventListener('mousedown', onClick)
    }
  }, [open])

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className={
          buttonClassName ||
          'flex items-center gap-1.5 rounded-full border border-line bg-surface px-3.5 py-1.5 text-sm font-medium hover:border-line-strong'
        }
      >
        {button}
      </button>
      {open && (
        <div
          className={`absolute z-40 mt-1 ${align === 'right' ? 'right-0' : 'left-0'} ${width} max-h-[75vh] overflow-y-auto rounded-xl border border-line bg-surface p-4 shadow-xl`}
        >
          {typeof children === 'function' ? children(() => setOpen(false)) : children}
        </div>
      )}
    </div>
  )
}

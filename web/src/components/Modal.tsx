import { useEffect, useRef, type ReactNode } from 'react'

// Minimal accessible modal: Esc closes, click on backdrop closes, focus is
// moved inside on open. Keep it dependency-free.
export function Modal({
  open,
  onClose,
  children,
  width = 'max-w-3xl',
  title
}: {
  open: boolean
  onClose: () => void
  children: ReactNode
  width?: string
  title?: ReactNode
}) {
  const ref = useRef<HTMLDivElement>(null)
  const onCloseRef = useRef(onClose)
  onCloseRef.current = onClose

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.stopPropagation()
        onCloseRef.current()
      }
    }
    // capture so page-level Esc handlers don't also fire
    window.addEventListener('keydown', onKey, true)
    // Move focus inside on open, but never steal it from an element already
    // inside the dialog (guards against any effect re-run / remount cause).
    // Prefer text inputs over the header close button.
    if (!ref.current?.contains(document.activeElement)) {
      const el =
        ref.current?.querySelector<HTMLElement>('input:not([type="checkbox"]), textarea, select, [autofocus]') ??
        ref.current?.querySelector<HTMLElement>('button, [tabindex]')
      el?.focus()
    }
    return () => window.removeEventListener('keydown', onKey, true)
  }, [open])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4 pt-[8vh]"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div
        ref={ref}
        role="dialog"
        aria-modal="true"
        className={`w-full ${width} rounded-2xl border border-line bg-surface shadow-2xl`}
      >
        {title !== undefined && (
          <div className="flex items-center justify-between border-b border-line px-5 py-3">
            <div className="text-base font-semibold">{title}</div>
            <button onClick={onClose} className="rounded p-1 text-muted hover:bg-surface-2 hover:text-fg" aria-label="Close">
              ✕
            </button>
          </div>
        )}
        <div className="p-5">{children}</div>
      </div>
    </div>
  )
}

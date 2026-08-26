import type { ReactNode } from 'react'

export function SectionCard({ title, children, actions }: { title: string; children: ReactNode; actions?: ReactNode }) {
  return (
    <section className="rounded-xl border border-line bg-surface p-4">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-base font-semibold">{title}</h2>
        {actions}
      </div>
      {children}
    </section>
  )
}

export function Field({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return (
    <label className="block">
      <span className="mb-1 block text-xs font-semibold uppercase tracking-wide text-muted">{label}</span>
      {children}
      {hint && <span className="mt-0.5 block text-xs text-muted">{hint}</span>}
    </label>
  )
}

export const inputCls =
  'w-full rounded-lg border border-line bg-surface-2 px-2 py-1.5 text-sm outline-none focus:border-accent'
export const btnCls = 'rounded-lg border border-line px-3 py-1.5 text-sm hover:bg-surface-2 disabled:opacity-50'
export const btnPrimaryCls = 'rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50'
export const btnDangerCls = 'rounded-lg border border-danger/40 px-3 py-1.5 text-sm text-danger hover:bg-danger/10'

export function SaveButton({
  onClick,
  pending,
  label = 'Save'
}: {
  onClick: () => void
  pending?: boolean
  label?: string
}) {
  return (
    <button onClick={onClick} disabled={pending} className={btnPrimaryCls}>
      {pending ? 'Saving…' : label}
    </button>
  )
}

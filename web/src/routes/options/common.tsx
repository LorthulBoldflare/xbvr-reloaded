import type { ReactNode } from 'react'

export function SectionCard({ title, children, actions }: { title: string; children: ReactNode; actions?: ReactNode }) {
  return (
    <section className="rounded-2xl border border-line bg-surface p-5 shadow-sm">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h2 className="text-base font-bold tracking-tight">{title}</h2>
        {actions}
      </div>
      {children}
    </section>
  )
}

export function Field({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return (
    <label className="block">
      <span className="mb-1 block text-[11px] font-bold uppercase tracking-wider text-muted">{label}</span>
      {children}
      {hint && <span className="mt-0.5 block text-xs text-muted">{hint}</span>}
    </label>
  )
}

export const inputCls =
  'w-full rounded-lg border border-line bg-surface-2 px-2.5 py-1.5 text-sm outline-none focus:border-accent'
export const btnCls =
  'rounded-full border border-line px-3.5 py-1.5 text-sm font-medium hover:bg-surface-2 disabled:opacity-50'
export const btnPrimaryCls =
  'btn-gradient rounded-full px-4 py-1.5 text-sm font-bold shadow disabled:opacity-50'
export const btnDangerCls =
  'rounded-full border border-danger/40 px-3.5 py-1.5 text-sm font-medium text-danger hover:bg-danger/10'

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

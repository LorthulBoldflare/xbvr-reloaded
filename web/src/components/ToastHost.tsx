import { useToastStore } from '../store/toasts'

const kindClasses = {
  info: 'border-line bg-surface text-fg',
  success: 'border-ok/40 bg-surface text-fg',
  error: 'border-danger/40 bg-surface text-fg'
}

const dotClasses = { info: 'bg-accent', success: 'bg-ok', error: 'bg-danger' }

export function ToastHost() {
  const toasts = useToastStore((s) => s.toasts)
  const dismiss = useToastStore((s) => s.dismiss)
  return (
    <div className="pointer-events-none fixed bottom-4 right-4 z-[100] flex w-96 flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`pointer-events-auto flex items-start gap-2 rounded-lg border px-3 py-2 shadow-lg ${kindClasses[t.kind]}`}
        >
          <span className={`mt-1.5 h-2 w-2 shrink-0 rounded-full ${dotClasses[t.kind]}`} />
          <span className="flex-1 text-sm">{t.message}</span>
          <button onClick={() => dismiss(t.id)} className="text-muted hover:text-fg" aria-label="Dismiss">
            ✕
          </button>
        </div>
      ))}
    </div>
  )
}

import { useUIStore } from '../store/ui'
import { Modal } from './Modal'

// Global confirm dialog (replaces $buefy.dialog.confirm).
export function ConfirmHost() {
  const confirm = useUIStore((s) => s.confirm)
  const closeConfirm = useUIStore((s) => s.closeConfirm)
  if (!confirm) return null
  return (
    <Modal open onClose={() => closeConfirm(false)} width="max-w-md" title={confirm.title}>
      {confirm.message && <p className="mb-4 whitespace-pre-line text-sm text-muted">{confirm.message}</p>}
      <div className="flex justify-end gap-2">
        <button
          onClick={() => closeConfirm(false)}
          className="rounded-lg border border-line px-3 py-1.5 text-sm hover:bg-surface-2"
        >
          Cancel
        </button>
        <button
          onClick={() => closeConfirm(true)}
          className={`rounded-lg px-3 py-1.5 text-sm text-white ${confirm.danger ? 'bg-danger' : 'bg-accent'}`}
        >
          {confirm.confirmLabel ?? 'Confirm'}
        </button>
      </div>
    </Modal>
  )
}

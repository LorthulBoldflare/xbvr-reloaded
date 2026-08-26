// Small toggle switch used across options and filters.
export function Toggle({
  checked,
  onChange,
  label,
  disabled
}: {
  checked: boolean
  onChange: (v: boolean) => void
  label?: string
  disabled?: boolean
}) {
  return (
    <label className={`flex cursor-pointer items-center gap-2 ${disabled ? 'opacity-50' : ''}`}>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        disabled={disabled}
        onClick={() => onChange(!checked)}
        className={`relative h-5 w-9 shrink-0 rounded-full transition-colors ${checked ? 'bg-accent' : 'bg-line-strong'}`}
      >
        <span
          className={`absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-transform ${checked ? 'translate-x-4.5 left-0 ml-0.5' : 'left-0.5'}`}
          style={{ transform: checked ? 'translateX(1.125rem)' : 'translateX(0.125rem)' }}
        />
      </button>
      {label && <span className="text-sm">{label}</span>}
    </label>
  )
}

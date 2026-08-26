// Read-only star rating display + optional interaction (0.5 steps, click to
// set, reset handled by caller).
export function StarRating({
  value,
  onChange,
  readonly = false,
  size = 'md'
}: {
  value: number
  onChange?: (v: number) => void
  readonly?: boolean
  size?: 'sm' | 'md' | 'lg'
}) {
  const px = size === 'lg' ? 22 : size === 'sm' ? 13 : 17
  const stars = [1, 2, 3, 4, 5]

  const star = (i: number) => {
    const fill = Math.max(0, Math.min(1, value - (i - 1)))
    return (
      <span key={i} className="relative inline-block" style={{ width: px, height: px }}>
        <svg viewBox="0 0 24 24" width={px} height={px} className="absolute text-line-strong">
          <path
            fill="currentColor"
            d="M12 17.27 18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"
          />
        </svg>
        <span className="absolute inset-0 overflow-hidden" style={{ width: `${fill * 100}%` }}>
          <svg viewBox="0 0 24 24" width={px} height={px} className="text-warn">
            <path
              fill="currentColor"
              d="M12 17.27 18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"
            />
          </svg>
        </span>
      </span>
    )
  }

  if (readonly) {
    return (
      <span className="inline-flex items-center" title={`${value} / 5`}>
        {stars.map(star)}
      </span>
    )
  }

  return (
    <span className="inline-flex items-center" title={`${value} / 5`}>
      {stars.map((i) => (
        <span key={i} className="relative inline-flex">
          {/* left half → .5, right half → full */}
          <button
            type="button"
            aria-label={`${i - 0.5} stars`}
            className="absolute left-0 top-0 z-10 h-full w-1/2 cursor-pointer"
            onClick={() => onChange?.(i - 0.5)}
          />
          <button
            type="button"
            aria-label={`${i} stars`}
            className="absolute right-0 top-0 z-10 h-full w-1/2 cursor-pointer"
            onClick={() => onChange?.(i)}
          />
          {star(i)}
        </span>
      ))}
    </span>
  )
}

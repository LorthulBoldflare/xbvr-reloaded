// Shared option matching for the tag pickers (TagFilterInput, TagInputEditor):
// case-insensitive substring, excluding already-selected names, capped at
// `limit` results.
export function filterTagOptions(options: string[], selected: ReadonlySet<string>, query: string, limit = 30): string[] {
  const q = query.toLowerCase()
  const matches: string[] = []
  for (const option of options) {
    if (!selected.has(option) && option.toLowerCase().includes(q)) matches.push(option)
    if (matches.length === limit) break
  }
  return matches
}

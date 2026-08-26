// Light/dark theme. Persisted in localStorage under 'xbvr-theme' (same key as
// the old UI); the initial value is applied pre-paint by an inline script in
// index.html.

export type Theme = 'light' | 'dark'

export function getTheme(): Theme {
  return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'
}

export function setTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme
  localStorage.setItem('xbvr-theme', theme)
}

export function toggleTheme(): Theme {
  const next = getTheme() === 'dark' ? 'light' : 'dark'
  setTheme(next)
  return next
}

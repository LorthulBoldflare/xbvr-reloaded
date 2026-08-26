// Theme management — light/dark with system-preference default and
// localStorage persistence. Applied to <html data-theme="…">, which the
// design tokens in assets/theme.css key off.

const STORAGE_KEY = 'xbvr-theme'
const THEMES = { light: '#f4f5f8', dark: '#0d1017' }

export function currentTheme () {
  return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light'
}

export function applyTheme (theme) {
  document.documentElement.setAttribute('data-theme', theme)
  document.documentElement.style.colorScheme = theme

  // theme-color should match the page background in the active theme
  let meta = document.querySelector('meta[name="theme-color"]')
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = 'theme-color'
    document.head.appendChild(meta)
  }
  meta.content = THEMES[theme] || THEMES.light

  try {
    localStorage.setItem(STORAGE_KEY, theme)
  } catch (e) {
    // private browsing etc. — theme just won't persist
  }
}

export function initTheme () {
  let theme = null
  try {
    theme = localStorage.getItem(STORAGE_KEY)
  } catch (e) {
    // ignore
  }
  if (theme !== 'dark' && theme !== 'light') {
    theme = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
  }
  applyTheme(theme)
  return theme
}

export function toggleTheme () {
  const next = currentTheme() === 'dark' ? 'light' : 'dark'
  applyTheme(next)
  return next
}

import { lazy, Suspense } from 'react'
import { createBrowserRouter } from 'react-router-dom'
import App from './App'

const ScenesPage = lazy(() => import('./routes/scenes/ScenesPage').then((m) => ({ default: m.ScenesPage })))
const ScenePage = lazy(() => import('./routes/scenes/ScenePage').then((m) => ({ default: m.ScenePage })))
const FilePage = lazy(() => import('./routes/files/FilePage').then((m) => ({ default: m.FilePage })))
const ActorsPage = lazy(() => import('./routes/actors/ActorsPage').then((m) => ({ default: m.ActorsPage })))
const OptionsPage = lazy(() => import('./routes/options/OptionsPage').then((m) => ({ default: m.OptionsPage })))

function route(element: JSX.Element) {
  return <Suspense fallback={<div className="py-16 text-center text-muted">Loading…</div>}>{element}</Suspense>
}

export const router = createBrowserRouter(
  [
    {
      path: '/',
      element: <App />,
      children: [
        { index: true, element: route(<ScenesPage />) },
        { path: 'scenes/:id', element: route(<ScenePage />) },
        { path: 'files/:id', element: route(<FilePage />) },
        { path: 'actors', element: route(<ActorsPage />) },
        { path: 'options/*', element: route(<OptionsPage />) }
      ]
    }
  ],
  { basename: import.meta.env.BASE_URL }
)

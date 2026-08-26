import { createBrowserRouter } from 'react-router-dom'
import App from './App'
import { ScenesPage } from './routes/scenes/ScenesPage'
import { ScenePage } from './routes/scenes/ScenePage'
import { FilePage } from './routes/files/FilePage'
import { ActorsPage } from './routes/actors/ActorsPage'
import { OptionsPage } from './routes/options/OptionsPage'

export const router = createBrowserRouter(
  [
    {
      path: '/',
      element: <App />,
      children: [
        { index: true, element: <ScenesPage /> },
        { path: 'scenes/:id', element: <ScenePage /> },
        { path: 'files/:id', element: <FilePage /> },
        { path: 'actors', element: <ActorsPage /> },
        { path: 'options/*', element: <OptionsPage /> }
      ]
    }
  ],
  { basename: import.meta.env.BASE_URL }
)

import { lazy, Suspense, useEffect } from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar, UpdateSnackbar } from './components/Sidebar'
import { MigrationOverlay } from './components/MigrationOverlay'
import { ToastHost } from './components/ToastHost'
import { ConfirmHost } from './components/ConfirmHost'
import { startSocket } from './ws/socket'
import { useUIStore } from './store/ui'

const QuickFind = lazy(() => import('./components/QuickFind').then((m) => ({ default: m.QuickFind })))
const SearchStashdbScenes = lazy(() =>
  import('./components/SearchStashdbScenes').then((m) => ({ default: m.SearchStashdbScenes }))
)
const ActorDetails = lazy(() => import('./components/ActorDetails').then((m) => ({ default: m.ActorDetails })))
const EditActor = lazy(() => import('./components/EditActor').then((m) => ({ default: m.EditActor })))
const SearchStashdbActors = lazy(() =>
  import('./components/SearchStashdbActors').then((m) => ({ default: m.SearchStashdbActors }))
)

export default function App() {
  const openQuickFind = useUIStore((s) => s.openQuickFind)
  const quickFindOpen = useUIStore((s) => s.quickFindOpen)
  const confirmOpen = useUIStore((s) => s.confirm !== null)
  const actorDetailsOpen = useUIStore((s) => s.actorDetailsId !== null)
  const editActorOpen = useUIStore((s) => s.editActorId !== null)
  const stashdbSceneOpen = useUIStore((s) => s.stashdbSceneSearchId !== null)
  const stashdbActorOpen = useUIStore((s) => s.stashdbActorSearchId !== null)
  const overlayOpen =
    quickFindOpen || confirmOpen || actorDetailsOpen || editActorOpen || stashdbSceneOpen || stashdbActorOpen

  useEffect(() => {
    startSocket()
  }, [])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      const typing = ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) || target.isContentEditable
      if (e.key === '?' && !typing && !overlayOpen) {
        e.preventDefault()
        openQuickFind(false)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [openQuickFind, overlayOpen])

  return (
    <div className="flex min-h-full">
      <Sidebar />
      <main className="min-w-0 flex-1 px-5 py-5 lg:px-8">
        <Outlet />
      </main>
      <UpdateSnackbar />
      <Suspense fallback={null}>
        {quickFindOpen && <QuickFind />}
        {stashdbSceneOpen && <SearchStashdbScenes />}
        {actorDetailsOpen && <ActorDetails />}
        {editActorOpen && <EditActor />}
        {stashdbActorOpen && <SearchStashdbActors />}
      </Suspense>
      <MigrationOverlay />
      <ConfirmHost />
      <ToastHost />
    </div>
  )
}

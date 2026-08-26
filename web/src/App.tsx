import { useEffect } from 'react'
import { Outlet } from 'react-router-dom'
import { Navbar } from './components/Navbar'
import { QuickFind } from './components/QuickFind'
import { MigrationOverlay } from './components/MigrationOverlay'
import { ToastHost } from './components/ToastHost'
import { ConfirmHost } from './components/ConfirmHost'
import { SearchStashdbScenes } from './components/SearchStashdbScenes'
import { ActorDetails } from './components/ActorDetails'
import { EditActor } from './components/EditActor'
import { SearchStashdbActors } from './components/SearchStashdbActors'
import { startSocket } from './ws/socket'
import { useUIStore } from './store/ui'

export default function App() {
  const openQuickFind = useUIStore((s) => s.openQuickFind)
  const quickFindOpen = useUIStore((s) => s.quickFindOpen)
  const confirmOpen = useUIStore((s) => s.confirm !== null)

  useEffect(() => {
    startSocket()
  }, [])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      const typing = ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) || target.isContentEditable
      if (e.key === '?' && !typing && !quickFindOpen && !confirmOpen) {
        e.preventDefault()
        openQuickFind(false)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [openQuickFind, quickFindOpen, confirmOpen])

  return (
    <div className="flex min-h-full flex-col">
      <Navbar />
      <main className="mx-auto w-full max-w-[1600px] flex-1 px-4 py-4">
        <Outlet />
      </main>
      <QuickFind />
      <SearchStashdbScenes />
      <ActorDetails />
      <EditActor />
      <SearchStashdbActors />
      <MigrationOverlay />
      <ConfirmHost />
      <ToastHost />
    </div>
  )
}

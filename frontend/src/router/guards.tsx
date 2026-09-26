import { Outlet } from 'react-router-dom'
import AuthGate from '../components/AuthGate'
import { VisibleChannelsProvider } from '../visibleChannels'

export function AdminGuard() {
  return (
    <AuthGate>
      <VisibleChannelsProvider>
        <Outlet />
      </VisibleChannelsProvider>
    </AuthGate>
  )
}

export function ConsoleGuard() {
  return <Outlet />
}

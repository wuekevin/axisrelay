import { Navigate, Outlet, useLocation } from 'react-router-dom'
import AuthGate from '../components/AuthGate'
import StateShell from '../components/StateShell'
import { useUserAuth } from '../providers/AppProviders'
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
  const auth = useUserAuth()
  const location = useLocation()
  if (auth.status === 'loading') {
    return <StateShell variant="page" loading>正在确认登录状态</StateShell>
  }
  if (!auth.authenticated) {
    return <Navigate to="/auth/login" replace state={{ from: location.pathname }} />
  }
  return <Outlet />
}

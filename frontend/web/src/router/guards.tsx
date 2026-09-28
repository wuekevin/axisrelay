import { Navigate, Outlet, useLocation } from 'react-router-dom'
import StateShell from '../components/StateShell'
import { useUserAuth } from '../providers/AppProviders'

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

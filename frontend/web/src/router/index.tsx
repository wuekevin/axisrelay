import { lazy } from 'react'
import { useRoutes } from 'react-router-dom'
import { authRoutes } from './auth.routes'
import { consoleRoutes } from './console.routes'
import { publicRoutes } from './public.routes'

const InstallPage = lazy(() => import('../pages/InstallPage'))

export default function AppRouter() {
  return useRoutes([
    { path: '/install', element: <InstallPage /> },
    ...publicRoutes,
    ...authRoutes,
    ...consoleRoutes,
  ])
}

import { useRoutes } from 'react-router-dom'
import { adminRoutes } from './admin.routes'
import { authRoutes } from './auth.routes'
import { consoleRoutes } from './console.routes'
import { publicRoutes } from './public.routes'

export default function AppRouter() {
  return useRoutes([
    ...publicRoutes,
    ...authRoutes,
    ...consoleRoutes,
    ...adminRoutes,
  ])
}
